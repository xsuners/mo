package chttp

import (
	"context"
	"fmt"

	"github.com/xsuners/mo/net/description"
)

// Options configure a Dial call. Options are set by the DialOption
// values passed to Dial.
type Options struct {
	unaryInt description.UnaryClientInterceptor
	// streamInt StreamClientInterceptor

	chainUnaryInts []description.UnaryClientInterceptor
	// chainStreamInts []StreamClientInterceptor

	credentials string `ini-name:"credentials" long:"natsc-credentials" description:"nats credentials"`
}

// Option configures how we set up the connection.
type Option func(*Options)

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

// Credentials returns a DialOption that specifies the interceptor for
// unary RPCs.
func Credentials(crets string) Option {
	return func(o *Options) {
		o.credentials = crets
	}
}

func defaultDialOptions() Options {
	return Options{}
}

// Conn impl description.UnaryClient for publish message to aim queue
type Conn struct {
	dopts Options
	// conf  *Config
	// conn *nats.Conn

	// client *Client
}

var _ description.UnaryClient = (*Conn)(nil)

// New .
func New(opts ...Option) (*Conn, error) {

	pub := &Conn{
		dopts: defaultDialOptions(),
	}

	for _, opt := range opts {
		opt(&pub.dopts)
	}

	chainUnaryClientInterceptors(pub)

	// // TODO
	// nc, err := nats.Connect(pub.dopts.urls, pub.dopts.nopts...)
	// if err != nil {
	// 	log.Fatalw("xnats: publisher connect error", "err", err)
	// 	return nil, err
	// }

	// pub.conn = nc
	return pub, nil
}

// chainUnaryClientInterceptors chains all unary client interceptors into one.
func chainUnaryClientInterceptors(cc *Conn) {
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

// // Close .
// func (pub *Publisher) Close() {
// 	pub.conn.Close()
// }

// Invoke .
// subject.service.method
func (pub *Conn) Invoke(ctx context.Context, sm string, args interface{}, reply interface{}, opts ...description.CallOption) error {
	if pub.dopts.unaryInt != nil {
		return pub.dopts.unaryInt(ctx, sm, args, reply, pub, invoke, opts...)
	}
	return invoke(ctx, sm, args, reply, pub, opts...)
}

func invoke(ctx context.Context, sm string, args interface{}, reply interface{}, cc description.UnaryClient, opts ...description.CallOption) error {

	// TODO
	pub, ok := cc.(*Conn)
	if !ok {
		return fmt.Errorf("xnats: pub invoke error: cc type (%T) not match", cc)
	}

	co := copool.Get().(*CallOptions)
	defer copool.Put(co)

	for _, o := range opts {
		o.Apply(co)
	}

	_ = pub

	// if sm != "" && sm[0] == '/' {
	// 	sm = sm[1:]
	// }
	// pos := strings.LastIndex(sm, "/")
	// if pos == -1 {
	// 	return fmt.Errorf("xnats: publisher use invalid method (%s) error", sm)
	// }
	// service := sm[:pos]
	// method := sm[pos+1:]

	// data, err := proto.Marshal(args.(proto.Message))
	// if err != nil {
	// 	return err
	// }

	// // TODO use sync.Pool
	// request := &message.Message{
	// 	Service: service,
	// 	Method:  method,
	// 	Data:    data,
	// }

	// md, ok := metadata.FromOutgoingContext(ctx)
	// if ok {
	// 	request.Metas = message.EncodeMetadata(md)
	// }

	// data, err = proto.Marshal(request)
	// if err != nil {
	// 	return err
	// }

	// if !co.WaitResponse { // pub-sub mode
	// 	if err := pub.conn.Publish(co.Subject, data); err != nil {
	// 		return err
	// 	}
	// 	return nil
	// }

	// msg, err := pub.conn.Request(co.Subject, data, co.Timeout)
	// if err != nil {
	// 	log.Errorwc(ctx, "invoke:Request", "subject", co.Subject, "err", err)
	// 	return err
	// }
	// response := &message.Message{} // TODO use sync.Pool
	// if err = proto.Unmarshal(msg.Data, response); err != nil {
	// 	return err
	// }
	// // TODO 讲错误信息包装成status返回
	// if response.Code != 0 {
	// 	return fmt.Errorf("xnats: response (%s) error", response.Desc)
	// }
	// err = proto.Unmarshal(response.Data, reply.(proto.Message))
	// return err
	return nil
}

// NewStream begins a streaming RPC.
func (pub *Conn) NewStream(ctx context.Context, desc *description.StreamDesc, method string, opts ...description.CallOption) (cs description.ClientStream, err error) {
	return
}
