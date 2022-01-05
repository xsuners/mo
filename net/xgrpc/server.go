package xgrpc

import (
	"context"
	"sync"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/encoding/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type Options struct {
	gopts  []grpc.ServerOption
	codecs []encoding.Codec
	// port   int
	// ip     string
}

var defaultOptions = Options{
	// port: 9000,
}

// Option sets server options.
type Option func(*Options)

// GRPCOption .
func GRPCOption(opts ...grpc.ServerOption) Option {
	return func(o *Options) {
		o.gopts = append(o.gopts, opts...)
	}
}

// UnaryInterceptor returns a Option that sets the UnaryServerInterceptor for the
// server. Only one unary interceptor can be installed. The construction of multiple
// interceptors (e.g., chaining) can be implemented at the caller.
func UnaryInterceptor(i description.UnaryServerInterceptor) Option {
	return func(o *Options) {
		o.gopts = append(o.gopts, grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			return i(ctx, req, &description.UnaryServerInfo{
				Server:     info.Server,
				FullMethod: info.FullMethod,
			}, description.UnaryHandler(handler))
		}))
	}
}

// UnaryInterceptor returns a Option that sets the UnaryServerInterceptor for the
// server. Only one unary interceptor can be installed. The construction of multiple
// interceptors (e.g., chaining) can be implemented at the caller.
func WithCodec(codec encoding.Codec) Option {
	return func(o *Options) {
		o.codecs = append(o.codecs, codec)
	}
}

// func Port(port int) Option {
// 	return func(o *options) {
// 		o.port = port
// 	}
// }

// // Config .
// type Config struct {
// 	IP   string `json:"ip"`
// 	Port int    `json:"port"`
// }

// Server .
type Server struct {
	// conf   *Config
	// lis      net.Listener
	*grpc.Server

	opts     Options
	mu       sync.Mutex
	services map[string]*description.ServiceInfo // origin
	checked  bool
}

// New .
func New(opt ...Option) (s *Server, cf func()) {
	s = &Server{
		opts:     defaultOptions,
		services: make(map[string]*description.ServiceInfo),
		// conf: c,
	}
	for _, o := range opt {
		o(&s.opts)
	}
	// register custom codec
	s.opts.codecs = append(s.opts.codecs, json.Codec{})
	for _, codec := range s.opts.codecs {
		encoding.RegisterCodec(codec)
	}
	// TODO opts
	s.Server = grpc.NewServer(s.opts.gopts...)
	cf = func() {
		log.Info("xgrpc is closing...")
		s.Stop()
		log.Info("xgrpc is closed.")
	}
	return
}

// // Server .
// func (s *Server) Server() *grpc.Server {
// 	return s.server
// }

// // Port .
// func (s *Server) Port() int {
// 	return s.opts.port
// }

// // Start .
// func (s *Server) Start(l net.Listener) (err error) {
// 	err = s.server.Serve(s.lis)
// 	if err != nil {
// 		log.Fatalf("failed to serve: %v", err)
// 	}
// 	return
// }

// RegisterService .
func (s *Server) RegisterService(sd *description.ServiceDesc, ss interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := description.Register(&s.services, sd, ss)
	if err != nil {
		log.Fatalw("xgrpc register service error", "err", err)
	}
	s.Server.RegisterService(convert(sd), ss)
}

// RegisterService .
func (s *Server) Register(ss interface{}, sds ...*description.ServiceDesc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sd := range sds {
		err := description.Register(&s.services, sd, ss)
		if err != nil {
			log.Fatalw("xgrpc register service error", "err", err)
		}
		s.Server.RegisterService(convert(sd), ss)
	}
	// TODO
	if s.checked {
		return
	}
	s.checked = true
	grpc_health_v1.RegisterHealthServer(s.Server, &Checker{})
}

// Service .
func (s *Server) Services() (services map[string]*description.ServiceInfo) {
	return s.services
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
func (s *Server) Stop() {
	s.Server.GracefulStop()
}
