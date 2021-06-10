package xnats

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/message"
	"github.com/xsuners/mo/net/util/ip"
	"github.com/xsuners/mo/net/util/unats"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// // Config .
// type Config struct {
// 	Subject     string `json:"subject"`
// 	Queue       string `json:"queue"`
// 	Credentials string `json:"credentials"`
// 	URLS        string `json:"urls"`
// 	Reply       bool   `json:"reply"`
// }

type consumerOptions struct {
	unaryInt       description.UnaryServerInterceptor
	chainUnaryInts []description.UnaryServerInterceptor
	nopts          []nats.Option
	queue          string
	credentials    string
	urls           string
	// subject        string
	// reply          bool
}

var defaultOptions = consumerOptions{
	urls: nats.DefaultURL,
}

// A Option sets options such as credentials, codec and keepalive parameters, etc.
type Option interface {
	apply(*consumerOptions)
}

// EmptyOption does not alter the server configuration. It can be embedded
// in another structure to build custom server options.
//
// Experimental
//
// Notice: This type is EXPERIMENTAL and may be changed or removed in a
// later release.
type EmptyOption struct{}

func (EmptyOption) apply(*consumerOptions) {}

// funcOption wraps a function that modifies consumerOptions into an
// implementation of the Option interface.
type funcOption struct {
	f func(*consumerOptions)
}

func (fdo *funcOption) apply(do *consumerOptions) {
	fdo.f(do)
}

func newFuncOption(f func(*consumerOptions)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// UnaryInterceptor returns a Option that sets the UnaryServerInterceptor for the
// server. Only one unary interceptor can be installed. The construction of multiple
// interceptors (e.g., chaining) can be implemented at the caller.
func UnaryInterceptor(i description.UnaryServerInterceptor) Option {
	return newFuncOption(func(o *consumerOptions) {
		if o.unaryInt != nil {
			panic("The unary server interceptor was already set and may not be reset.")
		}
		o.unaryInt = i
	})
}

// ChainUnaryInterceptor returns a Option that specifies the chained interceptor
// for unary RPCs. The first interceptor will be the outer most,
// while the last interceptor will be the inner most wrapper around the real call.
// All unary interceptors added by this method will be chained.
func ChainUnaryInterceptor(interceptors ...description.UnaryServerInterceptor) Option {
	return newFuncOption(func(o *consumerOptions) {
		o.chainUnaryInts = append(o.chainUnaryInts, interceptors...)
	})
}

// WithNatsOption config under nats .
func WithNatsOption(opt nats.Option) Option {
	return newFuncOption(func(o *consumerOptions) {
		o.nopts = append(o.nopts, opt)
	})
}

// // Subject .
// func Subject(subject string) Option {
// 	return newFuncOption(func(o *consumerOptions) {
// 		o.subject = subject
// 	})
// }

// Queue .
func Queue(queue string) Option {
	return newFuncOption(func(o *consumerOptions) {
		o.queue = queue
	})
}

// Credentials .
func Credentials(credentials string) Option {
	return newFuncOption(func(o *consumerOptions) {
		o.credentials = credentials
	})
}

// URLS .
func URLS(urls string) Option {
	return newFuncOption(func(o *consumerOptions) {
		o.urls = urls
	})
}

// // Reply .
// func Reply(reply bool) Option {
// 	return newFuncOption(func(o *consumerOptions) {
// 		o.reply = reply
// 	})
// }

// Server .
type Server struct {
	// conf     *Config
	opts     consumerOptions
	mu       sync.Mutex
	conn     *nats.Conn
	services map[string]*description.ServiceInfo // service name -> service info
}

// New .
func New(opt ...Option) (s *Server, cf func(), err error) {
	opts := defaultOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	s = &Server{
		opts:     opts,
		services: make(map[string]*description.ServiceInfo),
	}
	chainUnaryServerInterceptors(s)
	s.opts.nopts = setupConnOptions(s.opts.nopts)
	if s.opts.credentials != "" {
		s.opts.nopts = append(s.opts.nopts, nats.UserCredentials(s.opts.credentials))
	}
	conn, err := nats.Connect(s.opts.urls, s.opts.nopts...)
	if err != nil {
		return
	}
	s.conn = conn
	cf = func() {
		log.Info("xnats is closing...")
		s.Stop()
		log.Info("xnats is closed.")
	}
	return
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

// // Subscribe .
// func (c *Server) Subscribe() (err error) {
// 	if c.conf.Reply {
// 		c.conn.Subscribe(c.conf.Subject, c.processAndReply)
// 	} else {
// 		c.conn.Subscribe(c.conf.Subject, c.process)
// 	}
// 	c.conn.Flush()
// 	if err := c.conn.LastError(); err != nil {
// 		log.Error("xnats: get last error", zap.Error(err))
// 		return err
// 	}
// 	log.Infof("Listening on [%s]", c.conf.Subject)
// 	return
// }

// Serve .
func (c *Server) Serve() (err error) {
	// if c.conf.Reply {
	// 	c.conn.QueueSubscribe(c.conf.Subject, c.conf.Queue, c.processAndReply)
	// } else {
	// 	c.conn.QueueSubscribe(c.conf.Subject, c.conf.Queue, c.process)
	// }
	for subj := range c.services {
		c.conn.Subscribe("ip-"+subj+"."+unats.IPSubject(ip.Internal()), c.processAndReply)
		c.conn.QueueSubscribe(subj, c.opts.queue, c.processAndReply)
		c.conn.Subscribe("all-"+subj, c.processAndReply)
		log.Infow("xnats:start", "subject", subj)
	}
	c.conn.Flush()
	if err := c.conn.LastError(); err != nil {
		return err
	}
	return
}

// func (c *Server) process(msg *nats.Msg) {
// 	log.Debugw("xnats get a message", "subject", msg.Subject)

// 	// TODO add md
// 	ctx := context.TODO()
// 	in := &message.Message{}

// 	err := proto.Unmarshal(msg.Data, in)
// 	if err != nil {
// 		log.Errorsc(ctx, "xnats: unmarshal nats message error", zap.Error(err))
// 		return
// 	}

// 	srv, known := c.services[in.Service]
// 	if !known {
// 		log.Errorsc(ctx, "xnats: get service error", zap.String("service", in.Service))
// 		return
// 	}
// 	md, ok := srv.Method(in.Method)
// 	if !ok {
// 		log.Errorsc(ctx, "xnats: get method error", zap.String("method", in.Method))
// 		return
// 	}

// 	df := func(v interface{}) error {
// 		req, ok := v.(proto.Message)
// 		if !ok {
// 			return fmt.Errorf("in type %T is not proto.Message", v)
// 		}
// 		return proto.Unmarshal(in.Data, req)
// 	}
// 	if _, err = md.Handler(srv.Service(), ctx, df, nil); err != nil {
// 		log.Warnsc(ctx, "xnats: handle message error", zap.Error(err))
// 		return
// 	}
// }

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

// Stop .
func (c *Server) Stop() {
	c.conn.Drain()
}
