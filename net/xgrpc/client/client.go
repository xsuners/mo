package client

import (
	"context"
	"fmt"
	"strconv"
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

	ip       string
	port     int
	pkg      string
	service  string
	balancer string
	resolver string
}

// func (o *options) Value() interface{} {
// 	return o
// }

// DialOption .
type DialOption func(*options)

var defaultOptions = options{
	ip:       "127.0.0.1",
	port:     8500,
	balancer: ketama.Name,
}

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

// IP (e.g., chaining) can be implemented at the caller.
func IP(ip string) DialOption {
	return func(o *options) {
		o.ip = ip
	}
}

// Port (e.g., chaining) can be implemented at the caller.
func Port(port int) DialOption {
	return func(o *options) {
		o.port = port
	}
}

// Package .
func Package(pkg string) DialOption {
	return func(o *options) {
		o.pkg = pkg
	}
}

// Service (e.g., chaining) can be implemented at the caller.
func Service(service string) DialOption {
	return func(o *options) {
		o.service = service
	}
}

// Balancer (e.g., chaining) can be implemented at the caller.
func Balancer(balancer string) DialOption {
	return func(o *options) {
		o.balancer = balancer
	}
}

// Resolver .
func Resolver(resolver string) DialOption {
	return func(o *options) {
		o.resolver = resolver
	}
}

// // Config .
// type Config struct {
// 	IP       string `json:"ip"`
// 	Port     int    `json:"port"`
// 	Service  string `json:"service"`
// 	Balancer string `json:"balancer"`
// }

// Client .
type Client struct {
	opts   options
	target string
	// conf   *Config
	cc *grpc.ClientConn
}

var _ description.UnaryClient = (*Client)(nil)

// New .
func New(opt ...DialOption) (conn description.ClientConnInterface, err error) {
	client := &Client{
		opts: defaultOptions,
	}
	for _, o := range opt {
		o(&client.opts)
	}
	switch client.opts.resolver {
	case "consul":
		client.target = fmt.Sprintf("consul://%s:%d/%s", client.opts.ip, client.opts.port, client.opts.service)
	case "dns":
		client.target = fmt.Sprintf("dns:///%s.default.svc.cluster.local:9000", client.opts.pkg)
	default:
		client.target = client.opts.ip + ":" + strconv.Itoa(client.opts.port)
	}
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

	switch c.opts.balancer {
	case ketama.Name:
		log.Info("use balancer ketame")
		c.opts.gopts = append(c.opts.gopts, grpc.WithDefaultServiceConfig("{\"loadBalancingPolicy\":\""+ketama.Name+"\"}"))
	default:
		log.Info("use balancer round_robin")
		c.opts.gopts = append(c.opts.gopts, grpc.WithDefaultServiceConfig("{\"loadBalancingPolicy\":\""+rr.Name+"\"}"))
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	conn, err = grpc.DialContext(ctx, c.target, c.opts.gopts...)
	if err != nil {
		log.Errorw("grpc client dial error", "err", err, "target", c.target)
		return
	}

	return
}

// Invoke .
func (c *Client) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...description.CallOption) error {
	co := Options{}
	for _, o := range opts {
		o.Apply(&co)
	}
	return c.cc.Invoke(ctx, method, args, reply, co.copts...)
}

// NewStream begins a streaming RPC.
func (c *Client) NewStream(ctx context.Context, desc *description.StreamDesc, method string, opts ...description.CallOption) (cs description.ClientStream, err error) {
	co := Options{}
	for _, o := range opts {
		o.Apply(&co)
	}
	return c.cc.NewStream(ctx, &grpc.StreamDesc{
		StreamName: desc.StreamName,
		Handler: func(srv interface{}, stream grpc.ServerStream) error {
			return desc.Handler(srv, stream)
		},
		ServerStreams: desc.ServerStreams,
		ClientStreams: desc.ClientStreams,
	}, method, co.copts...)
}

// Close .
func (c *Client) Close() {
	c.cc.Close()
}

func init() {
	// register resolvers
	resolver.Register(consul.NewBuilder())
	// register balancers
	balancer.Register(ketama.Builder())
}
