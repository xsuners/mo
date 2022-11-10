package xgrpc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/misc/ip"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/encoding/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type Options struct {
	Port int `ini-name:"port" long:"grpc-port" description:"grpc port"`

	gopts  []grpc.ServerOption
	codecs []encoding.Codec
	ints   []grpc.UnaryServerInterceptor
}

var defaultOptions = Options{
	Port: 9000,
}

// Option sets server options.
type Option func(*Options)

// GRPCOption .
func GRPCOption(opts ...grpc.ServerOption) Option {
	return func(o *Options) {
		o.gopts = append(o.gopts, opts...)
	}
}

func UnaryInterceptor(ints ...description.UnaryServerInterceptor) Option {
	return func(o *Options) {
		for _, i := range ints {
			o.ints = append(o.ints, func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				return i(ctx, req, &description.UnaryServerInfo{
					Server:     info.Server,
					FullMethod: info.FullMethod,
				}, description.UnaryHandler(handler))
			})
		}
	}
}

func WithCodec(codec encoding.Codec) Option {
	return func(o *Options) {
		o.codecs = append(o.codecs, codec)
	}
}

func Port(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// Server .
type Server struct {
	*grpc.Server

	opts     Options
	mu       sync.Mutex
	services map[string]*description.ServiceInfo // origin
	checked  bool
}

// New .
func New(opt ...Option) (description.Server, func()) {
	s := &Server{
		opts:     defaultOptions,
		services: make(map[string]*description.ServiceInfo),
		// conf: c,
	}
	for _, o := range opt {
		o(&s.opts)
	}
	s.opts.gopts = append(s.opts.gopts, grpc.ChainUnaryInterceptor(s.opts.ints...))
	// register custom codec
	s.opts.codecs = append(s.opts.codecs, json.Codec{})
	for _, codec := range s.opts.codecs {
		encoding.RegisterCodec(codec)
	}
	// TODO opts
	s.Server = grpc.NewServer(s.opts.gopts...)
	return s, func() {
		log.Info("xgrpc is closing...")
		s.Stop()
		log.Info("xgrpc is closed.")
	}
}

func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.opts.Port))
	if err != nil {
		return err
	}
	s.opts.Port = lis.Addr().(*net.TCPAddr).Port
	return s.Server.Serve(lis)
}

func (s *Server) Naming(nm naming.Naming) error {
	services := make(map[string]struct{})
	for name := range s.services {
		services[strings.Split(name, ".")[0]] = struct{}{}
	}
	for service := range services {
		ins := &naming.Service{
			Name:     service,
			Protocol: naming.GRPC,
			IP:       ip.Internal(),
			Port:     s.opts.Port,
			Tag:      []string{"grpc"},
		}
		if err := nm.Register(ins); err != nil {
			return err
		}
	}
	return nil
}

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
