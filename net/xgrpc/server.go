package xgrpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/description"
	"google.golang.org/grpc"
)

type options struct {
	gopts []grpc.ServerOption
}

var defaultOptions = options{}

// Option sets server options.
type Option func(*options)

// GRPCOption .
func GRPCOption(opts ...grpc.ServerOption) Option {
	return func(o *options) {
		o.gopts = append(o.gopts, opts...)
	}
}

// UnaryInterceptor returns a Option that sets the UnaryServerInterceptor for the
// server. Only one unary interceptor can be installed. The construction of multiple
// interceptors (e.g., chaining) can be implemented at the caller.
func UnaryInterceptor(i description.UnaryServerInterceptor) Option {
	return func(o *options) {
		o.gopts = append(o.gopts, grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			return i(ctx, req, &description.UnaryServerInfo{
				Server:     info.Server,
				FullMethod: info.FullMethod,
			}, description.UnaryHandler(handler))
		}))
	}
}

// Config .
type Config struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// Server .
type Server struct {
	mu     sync.Mutex
	opts   options
	conf   *Config
	server *grpc.Server
	lis    net.Listener
}

// New .
func New(c *Config, opt ...Option) *Server {
	s := &Server{
		opts: defaultOptions,
		conf: c,
	}
	for _, o := range opt {
		o(&s.opts)
	}
	// TODO opts
	s.server = grpc.NewServer(s.opts.gopts...)
	return s
}

// Server .
func (s *Server) Server() *grpc.Server {
	return s.server
}

// Start .
func (s *Server) Start() (err error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.conf.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	err = s.server.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	return
}

// RegisterService .
func (s *Server) RegisterService(sd *description.ServiceDesc, ss interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.server.RegisterService(convert(sd), ss)
}

// convert .
func convert(in *description.ServiceDesc) (out *grpc.ServiceDesc) {
	out = &grpc.ServiceDesc{
		ServiceName: in.ServiceName,
		HandlerType: in.HandlerType,
		Metadata:    in.Metadata,
	}
	for _, m := range in.Methods {
		h := m.Handler
		out.Methods = append(out.Methods, grpc.MethodDesc{
			MethodName: m.MethodName,
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				return h(srv, ctx, dec, func(ctx context.Context, req interface{}, info *description.UnaryServerInfo, handler description.UnaryHandler) (resp interface{}, err error) {
					return interceptor(ctx, req, &grpc.UnaryServerInfo{
						Server:     info.Server,
						FullMethod: info.FullMethod,
					}, grpc.UnaryHandler(handler))
				})
			},
		})
	}
	return
}

// Stop .
func (s *Server) Stop(ctx context.Context) (err error) {
	return
}
