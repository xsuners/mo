package xcron

import (
	"context"
	"errors"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/leader_checker"
	"go.uber.org/zap"
)

type Options struct {
	unaryInt       description.UnaryServerInterceptor
	chainUnaryInts []description.UnaryServerInterceptor
	lc             leader_checker.Checker
}

var defaultOptions = Options{}

// Option sets server options.
type Option func(*Options)

func UnaryInterceptor(i description.UnaryServerInterceptor) Option {
	return func(o *Options) {
		if o.unaryInt != nil {
			panic("The unary server interceptor was already set and may not be reset.")
		}
		o.unaryInt = i
	}
}

func ChainUnaryInterceptor(interceptors ...description.UnaryServerInterceptor) Option {
	return func(o *Options) {
		o.chainUnaryInts = append(o.chainUnaryInts, interceptors...)
	}
}

func LC(lc leader_checker.Checker) Option {
	return func(o *Options) {
		o.lc = lc
	}
}

// Server .
type Server struct {
	cron     *cron.Cron
	opts     *Options
	mu       sync.Mutex
	services map[string]*description.ServiceInfo
}

// New .
func New(opt ...Option) (description.Server, func()) {
	opts := defaultOptions
	for _, opt := range opt {
		opt(&opts)
	}
	s := &Server{
		opts:     &opts,
		cron:     cron.New(cron.WithSeconds()),
		services: make(map[string]*description.ServiceInfo),
	}
	chainUnaryServerInterceptors(s)
	return s, func() {
		log.Info("xhttp is closing...")
		<-s.cron.Stop().Done()
		log.Info("xhttp is closed.")
	}
}

func chainUnaryServerInterceptors(s *Server) {
	interceptors := s.opts.chainUnaryInts
	if s.opts.unaryInt != nil {
		interceptors = append([]description.UnaryServerInterceptor{s.opts.unaryInt}, s.opts.chainUnaryInts...)
	}

	var chainedInt description.UnaryServerInterceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
		chainedInt = func(ctx context.Context, req interface{}, info *description.UnaryServerInfo, handler description.UnaryHandler) (interface{}, error) {
			return interceptors[0](ctx, req, info, getChainUnaryHandler(interceptors, 0, info, handler))
		}
	}

	s.opts.unaryInt = chainedInt
}

func getChainUnaryHandler(interceptors []description.UnaryServerInterceptor, curr int, info *description.UnaryServerInfo, finalHandler description.UnaryHandler) description.UnaryHandler {
	if curr == len(interceptors)-1 {
		return finalHandler
	}

	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return interceptors[curr+1](ctx, req, info, getChainUnaryHandler(interceptors, curr+1, info, finalHandler))
	}
}

// Register .
func (s *Server) Register(ss interface{}, sds ...*description.ServiceDesc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sd := range sds {
		err := description.Register(&s.services, sd, ss)
		if err != nil {
			log.Fatalw("xnats register service error", "err", err)
		}
	}
}

// Serve .
func (s *Server) Serve() (err error) {
	for _, desc := range s.services {
		for name, m := range desc.Methods() {
			if len(m.Cron) < 1 {
				m.Cron = "*/10 * * * * *"
			}
			if m.CheckLeader && s.opts.lc == nil {
				return errors.New("no leader checker supplied")
			}
			s.cron.AddFunc(m.Cron, func() {
				ctx := context.TODO()
				if m.CheckLeader && !s.opts.lc.IsLeader() {
					log.Infosc(ctx, "not leader")
					return
				}
				df := func(v interface{}) error {
					// req, ok := v.(proto.Message)
					// if !ok {
					// 	return fmt.Errorf("in type %T is not proto.Message", v)
					// }
					// return proto.Unmarshal(in.Data, req)
					return nil
				}
				_, err := m.Handler(desc.Service(), ctx, df, s.opts.unaryInt)
				if err != nil {
					log.Errorsc(ctx, name, zap.Error(err))
				}
			})
		}
	}
	s.cron.Start()
	return
}

func (s *Server) Naming(nm naming.Naming) error {
	return nil
}
