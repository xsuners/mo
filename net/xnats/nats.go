package xnats

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/misc/ip"
	"github.com/xsuners/mo/misc/unats"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/message"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type Options struct {
	Queue       string `ini-name:"queue" long:"nats-queue" description:"nats queue"`
	URLs        string `ini-name:"urls" long:"nats-urls" description:"nats urls"`
	Credentials string `ini-name:"credentials" long:"nats-credentials" description:"nats credentials"`

	unaryInt       description.UnaryServerInterceptor
	chainUnaryInts []description.UnaryServerInterceptor
	nopts          []nats.Option
}

var defaultOptions = Options{
	URLs: nats.DefaultURL,
}

// A Option sets options such as credentials, codec and keepalive parameters, etc.
type Option interface {
	apply(*Options)
}

type EmptyOption struct{}

func (EmptyOption) apply(*Options) {}

// funcOption wraps a function that modifies Option into an
// implementation of the Option interface.
type funcOption struct {
	f func(*Options)
}

func (fdo *funcOption) apply(do *Options) {
	fdo.f(do)
}

func newFuncOption(f func(*Options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// UnaryInterceptor .
func UnaryInterceptor(i description.UnaryServerInterceptor) Option {
	return newFuncOption(func(o *Options) {
		if o.unaryInt != nil {
			panic("The unary server interceptor was already set and may not be reset.")
		}
		o.unaryInt = i
	})
}

// ChainUnaryInterceptor .
func ChainUnaryInterceptor(interceptors ...description.UnaryServerInterceptor) Option {
	return newFuncOption(func(o *Options) {
		o.chainUnaryInts = append(o.chainUnaryInts, interceptors...)
	})
}

// WithNatsOption config under nats .
func WithNatsOption(opt nats.Option) Option {
	return newFuncOption(func(o *Options) {
		o.nopts = append(o.nopts, opt)
	})
}

// Queue .
func Queue(queue string) Option {
	return newFuncOption(func(o *Options) {
		o.Queue = queue
	})
}

// Credentials .
func Credentials(credentials string) Option {
	return newFuncOption(func(o *Options) {
		o.Credentials = credentials
	})
}

// URLS .
func URLS(urls string) Option {
	return newFuncOption(func(o *Options) {
		o.URLs = urls
	})
}

// Server .
type Server struct {
	opts     Options
	mu       sync.Mutex
	conn     *nats.Conn
	services map[string]*description.ServiceInfo
}

// New .
func New(opt ...Option) (description.Server, func(), error) {
	opts := defaultOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	s := &Server{
		opts:     opts,
		services: make(map[string]*description.ServiceInfo),
	}
	chainUnaryServerInterceptors(s)
	s.opts.nopts = setupConnOptions(s.opts.nopts)
	if s.opts.Credentials != "" {
		s.opts.nopts = append(s.opts.nopts, nats.UserCredentials(s.opts.Credentials))
	}
	log.Infos(s.opts.URLs)
	conn, err := nats.Connect(s.opts.URLs, s.opts.nopts...)
	if err != nil {
		return nil, nil, err
	}
	s.conn = conn
	return s, func() {
		log.Info("xnats is closing...")
		s.Stop()
		log.Info("xnats is closed.")
	}, nil
}

// chainUnaryServerInterceptors chains all unary server interceptors into one.
func chainUnaryServerInterceptors(s *Server) {
	// Prepend opts.unaryInt to the chaining interceptors if it exists, since unaryInt will
	// be executed before any other chained interceptors.
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

// getChainUnaryHandler recursively generate the chained UnaryHandler
func getChainUnaryHandler(interceptors []description.UnaryServerInterceptor, curr int, info *description.UnaryServerInfo, finalHandler description.UnaryHandler) description.UnaryHandler {
	if curr == len(interceptors)-1 {
		return finalHandler
	}

	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return interceptors[curr+1](ctx, req, info, getChainUnaryHandler(interceptors, curr+1, info, finalHandler))
	}
}

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.Name("NATS Sample Responder"))
	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))

	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		log.Infof("Disconnected due to: %s, will attempt reconnects for %.0fm", err, totalWait.Minutes())
	}))

	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Infof("Reconnected [%s]", nc.ConnectedUrl())
	}))

	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Infof("Exiting: %v", nc.LastError())
	}))

	return opts
}

var _ description.ServiceRegistrar = (*Server)(nil)

// RegisterService .
func (c *Server) RegisterService(sd *description.ServiceDesc, ss interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := description.Register(&c.services, sd, ss)
	if err != nil {
		log.Fatalw("xnats register service error", "err", err)
	}
}

// Register .
func (c *Server) Register(ss interface{}, sds ...*description.ServiceDesc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, sd := range sds {
		err := description.Register(&c.services, sd, ss)
		if err != nil {
			log.Fatalw("xnats register service error", "err", err)
		}
	}
}

// Serve .
func (c *Server) Serve() (err error) {
	for subj := range c.services {
		c.conn.Subscribe("ip-"+subj+"."+unats.IPSubject(ip.Internal()), c.processAndReply)
		c.conn.QueueSubscribe(subj, c.opts.Queue, c.processAndReply)
		c.conn.Subscribe("all-"+subj, c.processAndReply)
		log.Infow("xnats:start", "subject", subj)
	}
	c.conn.Flush()
	if err := c.conn.LastError(); err != nil {
		return err
	}
	return
}

func (c *Server) processAndReply(msg *nats.Msg) {
	log.Debugw("xnats get a message", "subject", msg.Subject)
	ctx := context.Background()
	in := &message.Message{}
	err := proto.Unmarshal(msg.Data, in)
	if err != nil {
		desc := "xnats unmarshal nats message error"
		reply(ctx, msg, 1, desc, nil)
		return
	}
	nmd := message.DecodeMetadata(in.Metas)
	ctx = metadata.NewIncomingContext(ctx, nmd)
	srv, known := c.services[in.Service]
	if !known {
		desc := fmt.Sprintf("xnats get service (%s) error", in.Service)
		reply(ctx, msg, 1, desc, nil)
		return
	}
	md, ok := srv.Method(in.Method)
	if !ok {
		desc := fmt.Sprintf("xnats get method (%s) error", in.Method)
		reply(ctx, msg, 1, desc, nil)
		return
	}
	df := func(v interface{}) error {
		req, ok := v.(proto.Message)
		if !ok {
			return fmt.Errorf("in type %T is not proto.Message", v)
		}
		return proto.Unmarshal(in.Data, req)
	}
	out, err := md.Handler(srv.Service(), ctx, df, c.opts.unaryInt)
	if err != nil {
		reply(ctx, msg, 1, err.Error(), nil)
		return
	}
	om, ok := out.(proto.Message)
	if !ok {
		desc := fmt.Sprintf("xnats out message (%T) not proto.Message", out)
		reply(ctx, msg, 1, desc, nil)
		return
	}
	data, err := proto.Marshal(om)
	if err != nil {
		desc := "xnats marshal out message error"
		reply(ctx, msg, 1, desc, nil)
		return
	}
	reply(ctx, msg, 0, "", data)
}

func reply(ctx context.Context, msg *nats.Msg, code int32, desc string, data []byte) {
	if msg.Reply == "" {
		log.Debugc(ctx, "reply:without reply")
		return
	}
	response := &message.Message{}
	if code != 0 {
		response.Code = code
		response.Desc = desc
		log.Errorc(ctx, desc)
	} else if data == nil {
		response.Code = code
		response.Desc = "xnats internal error, code id 0 but data is nil"
		log.Errorc(ctx, desc)
	} else {
		response.Data = data
	}
	data, err := proto.Marshal(response)
	if err != nil {
		log.Errorwc(ctx, "xnats marshal response error", "err", err)
		return
	}
	if err = msg.Respond(data); err != nil {
		log.Errorwc(ctx, "xnats response error", "err", err)
	}
}

func (s *Server) Naming(nm naming.Naming) error {
	return nil
}

// Stop .
func (c *Server) Stop() {
	c.conn.Drain()
}
