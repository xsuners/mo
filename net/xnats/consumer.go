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
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// Config .
type Config struct {
	Subject     string `json:"subject"`
	Queue       string `json:"queue"`
	Credentials string `json:"credentials"`
	URLS        string `json:"urls"`
	Reply       bool   `json:"reply"`
}

type consumerOptions struct {
	unaryInt       description.UnaryServerInterceptor
	chainUnaryInts []description.UnaryServerInterceptor

	nopts []nats.Option
}

var defaultConsumerOptions = consumerOptions{}

// A ConsumerOption sets options such as credentials, codec and keepalive parameters, etc.
type ConsumerOption interface {
	apply(*consumerOptions)
}

// EmptyConsumerOption does not alter the server configuration. It can be embedded
// in another structure to build custom server options.
//
// Experimental
//
// Notice: This type is EXPERIMENTAL and may be changed or removed in a
// later release.
type EmptyConsumerOption struct{}

func (EmptyConsumerOption) apply(*consumerOptions) {}

// funcConsumerOption wraps a function that modifies consumerOptions into an
// implementation of the ConsumerOption interface.
type funcConsumerOption struct {
	f func(*consumerOptions)
}

func (fdo *funcConsumerOption) apply(do *consumerOptions) {
	fdo.f(do)
}

func newFuncConsumerOption(f func(*consumerOptions)) *funcConsumerOption {
	return &funcConsumerOption{
		f: f,
	}
}

// UnaryInterceptor returns a ConsumerOption that sets the UnaryServerInterceptor for the
// server. Only one unary interceptor can be installed. The construction of multiple
// interceptors (e.g., chaining) can be implemented at the caller.
func UnaryInterceptor(i description.UnaryServerInterceptor) ConsumerOption {
	return newFuncConsumerOption(func(o *consumerOptions) {
		if o.unaryInt != nil {
			panic("The unary server interceptor was already set and may not be reset.")
		}
		o.unaryInt = i
	})
}

// ChainUnaryInterceptor returns a ConsumerOption that specifies the chained interceptor
// for unary RPCs. The first interceptor will be the outer most,
// while the last interceptor will be the inner most wrapper around the real call.
// All unary interceptors added by this method will be chained.
func ChainUnaryInterceptor(interceptors ...description.UnaryServerInterceptor) ConsumerOption {
	return newFuncConsumerOption(func(o *consumerOptions) {
		o.chainUnaryInts = append(o.chainUnaryInts, interceptors...)
	})
}

// WithNatsOption config under nats .
func WithNatsOption(opt nats.Option) ConsumerOption {
	return newFuncConsumerOption(func(o *consumerOptions) {
		o.nopts = append(o.nopts, opt)
	})
}

// Consumer .
type Consumer struct {
	opts     consumerOptions
	mu       sync.Mutex
	conf     *Config
	conn     *nats.Conn
	services map[string]*description.ServiceInfo // service name -> service info
}

// NewConsumer .
func NewConsumer(c *Config, opt ...ConsumerOption) *Consumer {
	opts := defaultConsumerOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	s := &Consumer{
		opts:     opts,
		conf:     c,
		services: make(map[string]*description.ServiceInfo),
	}
	chainUnaryServerInterceptors(s)

	// Connect Options.
	// opts := []nats.Option{nats.Name("NATS Sample Responder")}
	s.opts.nopts = append(s.opts.nopts, nats.Name("NATS Sample Responder"))
	s.opts.nopts = setupConnOptions(s.opts.nopts)

	// Use UserCredentials
	if c.Credentials != "" {
		s.opts.nopts = append(s.opts.nopts, nats.UserCredentials(c.Credentials))
	}

	// Connect to NATS
	conn, err := nats.Connect(c.URLS, s.opts.nopts...)
	if err != nil {
		log.Fatal(err)
	}

	s.conn = conn

	return s
}

// chainUnaryServerInterceptors chains all unary server interceptors into one.
func chainUnaryServerInterceptors(s *Consumer) {
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

var _ description.ServiceRegistrar = (*Consumer)(nil)

// RegisterService .
func (c *Consumer) RegisterService(sd *description.ServiceDesc, ss interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := description.Register(&c.services, sd, ss)
	if err != nil {
		log.Fatalw("xnats register service error", "err", err)
	}
}

// Subscribe .
func (c *Consumer) Subscribe() (err error) {
	if c.conf.Reply {
		c.conn.Subscribe(c.conf.Subject, c.processAndReply)
	} else {
		c.conn.Subscribe(c.conf.Subject, c.process)
	}
	c.conn.Flush()
	if err := c.conn.LastError(); err != nil {
		log.Error("xnats: get last error", zap.Error(err))
		return err
	}
	log.Infof("Listening on [%s]", c.conf.Subject)
	return
}

// QueueSubscribe .
func (c *Consumer) QueueSubscribe() (err error) {
	if c.conf.Reply {
		c.conn.QueueSubscribe(c.conf.Subject, c.conf.Queue, c.processAndReply)
	} else {
		c.conn.QueueSubscribe(c.conf.Subject, c.conf.Queue, c.process)
	}
	c.conn.Flush()
	if err := c.conn.LastError(); err != nil {
		return err
	}
	log.Infof("xnats: listening on [%s]", c.conf.Subject)
	return
}

func (c *Consumer) process(msg *nats.Msg) {
	log.Debugw("xnats get a message", "subject", msg.Subject)

	// TODO add md
	ctx := context.TODO()
	in := &message.Message{}

	err := proto.Unmarshal(msg.Data, in)
	if err != nil {
		log.Errorsc(ctx, "xnats: unmarshal nats message error", zap.Error(err))
		return
	}

	srv, known := c.services[in.Service]
	if !known {
		log.Errorsc(ctx, "xnats: get service error", zap.String("service", in.Service))
		return
	}
	md, ok := srv.Method(in.Method)
	if !ok {
		log.Errorsc(ctx, "xnats: get method error", zap.String("method", in.Method))
		return
	}

	df := func(v interface{}) error {
		req, ok := v.(proto.Message)
		if !ok {
			return fmt.Errorf("in type %T is not proto.Message", v)
		}
		return proto.Unmarshal(in.Data, req)
	}
	if _, err = md.Handler(srv.Service(), ctx, df, nil); err != nil {
		log.Warnsc(ctx, "xnats: handle message error", zap.Error(err))
		return
	}
}

func (c *Consumer) processAndReply(msg *nats.Msg) {
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
func (c *Consumer) Stop(ctx context.Context) (err error) {
	c.conn.Drain()
	return
}
