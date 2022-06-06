package xcron

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/description"
)

type Options struct {
	unaryInt       description.UnaryServerInterceptor
	chainUnaryInts []description.UnaryServerInterceptor
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

// Server .
type Server struct {
	*gin.Engine

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
		Engine:   gin.Default(),
		services: make(map[string]*description.ServiceInfo),
	}
	chainUnaryServerInterceptors(s)
	return s, func() {
		log.Info("xhttp is closing...")
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

func (s *Server) Check(c *gin.Context) {}

// Serve .
func (s *Server) Serve() (err error) {
	return
}

func (s *Server) Naming(nm naming.Naming) error {
	return nil
}
