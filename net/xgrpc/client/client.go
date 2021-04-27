package client

import (
	"context"
	"fmt"
	"time"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/xgrpc/client/balancer/ketama"
	"github.com/xsuners/mo/net/xgrpc/client/balancer/rr"
	"github.com/xsuners/mo/net/xgrpc/client/resolver/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

type options struct {
	gopts []grpc.DialOption
}

// func (o *options) Value() interface{} {
// 	return o
// }

// DialOption .
type DialOption func(*options)

var defaultOptions = options{}

// GRPCOption .
func GRPCOption(opts ...grpc.DialOption) DialOption {
	return func(o *options) {
		o.gopts = append(o.gopts, opts...)
	}
}

// UnaryInterceptor returns a DialOption that sets the UnaryServerInterceptor for the
// server. Only one  unary interceptor can be installed. The construction of multiple
// interceptors (e.g., chaining) can be implemented at the caller.
func UnaryInterceptor(i description.UnaryClientInterceptor) DialOption {
	return func(o *options) {
		// TODO 处理cc和callOption
		o.gopts = append(o.gopts, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return i(ctx, method, req, reply, nil, func(ctx context.Context, method string, req, reply interface{}, icc description.UnaryClient, iopts ...description.CallOption) error {
				return invoker(ctx, method, req, reply, cc, opts...)
			})
		}))
	}
}

// Config .
type Config struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Service  string `json:"service"`
	Balancer string `json:"balancer"`
}

// Client .
type Client struct {
	opts   options
	target string
	conf   *Config
	cc     *grpc.ClientConn
}

var _ description.UnaryClient = (*Client)(nil)

// New .
func New(c *Config, opt ...DialOption) (conn description.ClientConnInterface, err error) {
	client := &Client{
		opts: defaultOptions,
	}
	for _, o := range opt {
		o(&client.opts)
	}
	client.conf = c
	// client.opts = append(client.opts, opts...)
	client.target = fmt.Sprintf("consul://%s:%d/%s", c.IP, c.Port, c.Service)
	client.cc, err = client.dial()
	if err != nil {
		return
	}
	return client, nil
}

// dial .
func (c *Client) dial() (conn *grpc.ClientConn, err error) {
	// c.opts = append(c.opts, grpc.WithBlock(), grpc.WithInsecure())
	c.opts.gopts = append(c.opts.gopts, grpc.WithInsecure())

	switch c.conf.Balancer {
	case ketama.Name:
		log.Info("use balancer ketame")
		c.opts.gopts = append(c.opts.gopts, grpc.WithBalancerName(ketama.Name))
	default:
		log.Info("use balancer round_robin")
		c.opts.gopts = append(c.opts.gopts, grpc.WithBalancerName(rr.Name))
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	conn, err = grpc.DialContext(ctx, c.target, c.opts.gopts...)
	if err != nil {
		log.Errorw("grpc client dial error", "err", err, "target", c.target)
	}

	return
}

// Invoke .
func (c *Client) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...description.CallOption) error {
	// TODO 完成CallOption的抽象
	return c.cc.Invoke(ctx, method, args, reply)
}

// NewStream begins a streaming RPC.
func (c *Client) NewStream(ctx context.Context, desc *description.StreamDesc, method string, opts ...description.CallOption) (cs description.ClientStream, err error) {
	return
}

// Close .
func (c *Client) Close() {
	c.cc.Close()
}

func init() {
	// register resolvers
	resolver.Register(consul.Builder())
	// register balancers
	balancer.Register(ketama.Builder())
}
