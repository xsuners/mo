package xgrpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/util/ip"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

type options struct {
	gopts  []grpc.ServerOption
	codecs []encoding.Codec
	port   int
	// ip     string
}

var defaultOptions = options{
	port: 9000,
}

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

// UnaryInterceptor returns a Option that sets the UnaryServerInterceptor for the
// server. Only one unary interceptor can be installed. The construction of multiple
// interceptors (e.g., chaining) can be implemented at the caller.
func WithCodec(codec encoding.Codec) Option {
	return func(o *options) {
		o.codecs = append(o.codecs, codec)
	}
}

func Port(port int) Option {
	return func(o *options) {
		o.port = port
	}
}

// // Config .
// type Config struct {
// 	IP   string `json:"ip"`
// 	Port int    `json:"port"`
// }

// Server .
type Server struct {
	// conf   *Config
	mu     sync.Mutex
	opts   options
	server *grpc.Server
	lis    net.Listener
}

// New .
func New(opt ...Option) (s *Server, cf func()) {
	s = &Server{
		opts: defaultOptions,
		// conf: c,
	}
	for _, o := range opt {
		o(&s.opts)
	}
	// register custom codec
	for _, codec := range s.opts.codecs {
		encoding.RegisterCodec(codec)
	}
	// TODO opts
	s.server = grpc.NewServer(s.opts.gopts...)
	cf = func() {
		log.Info("xgrpc is closing...")
		s.Stop()
		log.Info("xgrpc is closed.")
	}
	return
}

// Server .
func (s *Server) Server() *grpc.Server {
	return s.server
}

// Port .
func (s *Server) Port() int {
	return s.opts.port
}

// Start .
func (s *Server) Start(n *naming.Naming) (err error) {
	s.lis, err = net.Listen("tcp", fmt.Sprintf(":%d", s.opts.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// register service to consul
	for service := range s.server.GetServiceInfo() {
		ins := &naming.Service{
			Name: service,
			IP:   ip.Internal(),
			Port: s.opts.port,
			Tag:  []string{"grpc"},
		}
		if err = n.Regitser(ins); err != nil {
			log.Errorw("xtcp: register service error", "err", err)
			return
		}
		log.Infow("xtcp: register service success", "service", ins.Name)
	}
	err = s.server.Serve(s.lis)
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
func (s *Server) Stop() {
	s.server.GracefulStop()
}
