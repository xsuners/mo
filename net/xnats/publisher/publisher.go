package publisher

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/message"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// // Config is used to config publisher
// type Config struct {
// 	Credentials    string        `json:"credentials"`
// 	URLS           string        `json:"urls"`
// 	Timeout        time.Duration `json:"timeout"`
// 	DefaultSubject string        `json:"default_subject"`
// }

// Options configure a Dial call. Options are set by the DialOption
// values passed to Dial.
type Options struct {
	unaryInt description.UnaryClientInterceptor
	// streamInt StreamClientInterceptor

	chainUnaryInts []description.UnaryClientInterceptor
	// chainStreamInts []StreamClientInterceptor

	nopts []nats.Option

	credentials    string        `ini-name:"credentials" long:"natsc-credentials" description:"nats credentials"`
	urls           string        `ini-name:"urls" long:"natsc-urls" description:"nats urls"`
	defaultTimeout time.Duration `ini-name:"defaultTimeout" long:"natsc-default-timeout" description:"nats defaultTimeout"`
	defaultSubject string        `ini-name:"defaultSubject" long:"natsc-default-subject" description:"nats defaultSubject"`
}

// Option configures how we set up the connection.
type Option func(*Options)

// EmptyDialOption does not alter the dial configuration. It can be embedded in
// another structure to build custom dial options.
//
// Experimental
//
// Notice: This type is EXPERIMENTAL and may be changed or removed in a
// later release.
// type EmptyDialOption struct{}

// func (EmptyDialOption) apply(*dialOptions) {}

// // funcDialOption wraps a function that modifies dialOptions into an
// // implementation of the DialOption interface.
// type funcDialOption struct {
// 	f func(*dialOptions)
// }

// func (fdo *funcDialOption) apply(do *dialOptions) {
// 	fdo.f(do)
// }

// func newFuncDialOption(f func(*dialOptions)) *funcDialOption {
// 	return &funcDialOption{
// 		f: f,
// 	}
// }

// WithUnaryInterceptor returns a DialOption that specifies the interceptor for
// unary RPCs.
func WithUnaryInterceptor(f description.UnaryClientInterceptor) Option {
	return func(o *Options) {
		o.unaryInt = f
	}
}

// WithChainUnaryInterceptor returns a DialOption that specifies the chained
// interceptor for unary RPCs. The first interceptor will be the outer most,
// while the last interceptor will be the inner most wrapper around the real call.
// All interceptors added by this method will be chained, and the interceptor
// defined by WithUnaryInterceptor will always be prepended to the chain.
func WithChainUnaryInterceptor(interceptors ...description.UnaryClientInterceptor) Option {
	return func(o *Options) {
		o.chainUnaryInts = append(o.chainUnaryInts, interceptors...)
	}
}

// WithNatsOption config under nats .
func WithNatsOption(opt nats.Option) Option {
	return func(o *Options) {
		o.nopts = append(o.nopts, opt)
	}
}

// Credentials returns a DialOption that specifies the interceptor for
// unary RPCs.
func Credentials(crets string) Option {
	return func(o *Options) {
		o.credentials = crets
	}
}

// URLS returns a DialOption that specifies the interceptor for
// unary RPCs.
func URLS(urls string) Option {
	return func(o *Options) {
		o.urls = urls
	}
}

// DefaultTimeout returns a DialOption that specifies the interceptor for
// unary RPCs.
func DefaultTimeout(du time.Duration) Option {
	return func(o *Options) {
		o.defaultTimeout = du
	}
}

// DefaultSubject returns a DialOption that specifies the interceptor for
// unary RPCs.
func DefaultSubject(subject string) Option {
	return func(o *Options) {
		o.defaultSubject = subject
	}
}

func defaultDialOptions() Options {
	return Options{
		defaultTimeout: time.Second * 2,
		// disableRetry:    !envconfig.Retry,
		// healthCheckFunc: internal.HealthCheckFunc,
		// copts: transport.ConnectOptions{
		// 	WriteBufferSize: defaultWriteBufSize,
		// 	ReadBufferSize:  defaultReadBufSize,
		// },
		// resolveNowBackoff: internalbackoff.DefaultExponential.Backoff,
		// withProxy:         true,
	}
}

// Publisher impl description.UnaryClient for publish message to aim queue
type Publisher struct {
	dopts Options
	// conf  *Config
	conn *nats.Conn
}

var _ description.ClientConnInterface = (*Publisher)(nil)

// New .
func New(opts ...Option) (*Publisher, error) {

	pub := &Publisher{
		dopts: defaultDialOptions(),
		// conf:  c,
	}
	// pub.fixConfig()

	for _, opt := range opts {
		opt(&pub.dopts)
	}

	chainUnaryClientInterceptors(pub)

	// TODO
	nc, err := nats.Connect(pub.dopts.urls, pub.dopts.nopts...)
	if err != nil {
		log.Fatalw("xnats: publisher connect error", "err", err)
		return nil, err
	}

	pub.conn = nc
	return pub, nil
}

// func (pub *Publisher) fixConfig() {
// 	if pub.conf == nil {
// 		pub.conf = &Config{}
// 	}
// 	if pub.conf.Timeout == 0 {
// 		pub.conf.Timeout = time.Second * 2
// 	}
// }

// chainUnaryClientInterceptors chains all unary client interceptors into one.
func chainUnaryClientInterceptors(cc *Publisher) {
	interceptors := cc.dopts.chainUnaryInts
	// Prepend dopts.unaryInt to the chaining interceptors if it exists, since unaryInt will
	// be executed before any other chained interceptors.
	if cc.dopts.unaryInt != nil {
		interceptors = append([]description.UnaryClientInterceptor{cc.dopts.unaryInt}, interceptors...)
	}
	var chainedInt description.UnaryClientInterceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
		chainedInt = func(ctx context.Context, method string, req, reply interface{}, cc description.UnaryClient, invoker description.UnaryInvoker, opts ...description.CallOption) error {
			return interceptors[0](ctx, method, req, reply, cc, getChainUnaryInvoker(interceptors, 0, invoker), opts...)
		}
	}
	cc.dopts.unaryInt = chainedInt
}

// getChainUnaryInvoker recursively generate the chained unary invoker.
func getChainUnaryInvoker(interceptors []description.UnaryClientInterceptor, curr int, finalInvoker description.UnaryInvoker) description.UnaryInvoker {
	if curr == len(interceptors)-1 {
		return finalInvoker
	}
	return func(ctx context.Context, method string, req, reply interface{}, cc description.UnaryClient, opts ...description.CallOption) error {
		return interceptors[curr+1](ctx, method, req, reply, cc, getChainUnaryInvoker(interceptors, curr+1, finalInvoker), opts...)
	}
}

// Close .
func (pub *Publisher) Close() {
	pub.conn.Flush()
	pub.conn.Close()
}

// Invoke .
// subject.service.method
func (pub *Publisher) Invoke(ctx context.Context, sm string, args interface{}, reply interface{}, opts ...description.CallOption) error {
	if pub.dopts.unaryInt != nil {
		return pub.dopts.unaryInt(ctx, sm, args, reply, pub, invoke, opts...)
	}
	return invoke(ctx, sm, args, reply, pub, opts...)
}

func invoke(ctx context.Context, sm string, args interface{}, reply interface{}, cc description.UnaryClient, opts ...description.CallOption) error {

	// TODO
	pub, ok := cc.(*Publisher)
	if !ok {
		return fmt.Errorf("xnats: pub invoke error: cc type (%T) not match", cc)
	}

	co := copool.Get().(*CallOptions)
	defer copool.Put(co)

	co.Timeout = pub.dopts.defaultTimeout
	co.Subject = pub.dopts.defaultSubject
	for _, o := range opts {
		o.Apply(co)
	}

	if sm != "" && sm[0] == '/' {
		sm = sm[1:]
	}
	pos := strings.LastIndex(sm, "/")
	if pos == -1 {
		return fmt.Errorf("xnats: publisher use invalid method (%s) error", sm)
	}
	service := sm[:pos]
	method := sm[pos+1:]

	data, err := proto.Marshal(args.(proto.Message))
	if err != nil {
		return err
	}

	// TODO use sync.Pool
	request := &message.Message{
		Service: service,
		Method:  method,
		Data:    data,
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		request.Metas = message.EncodeMetadata(md)
	}

	data, err = proto.Marshal(request)
	if err != nil {
		return err
	}

	if !co.WaitResponse { // pub-sub mode
		if err := pub.conn.Publish(co.Subject, data); err != nil {
			return err
		}
		return nil
	}

	msg, err := pub.conn.Request(co.Subject, data, co.Timeout)
	if err != nil {
		log.Errorwc(ctx, "invoke:Request", "subject", co.Subject, "err", err)
		return err
	}
	response := &message.Message{} // TODO use sync.Pool
	if err = proto.Unmarshal(msg.Data, response); err != nil {
		return err
	}
	// TODO 讲错误信息包装成status返回
	if response.Code != 0 {
		return fmt.Errorf("xnats: response (%s) error", response.Desc)
	}
	err = proto.Unmarshal(response.Data, reply.(proto.Message))
	return err
}

// NewStream begins a streaming RPC.
// TODO
func (pub *Publisher) NewStream(ctx context.Context, desc *description.StreamDesc, method string, opts ...description.CallOption) (cs description.ClientStream, err error) {
	return
}
