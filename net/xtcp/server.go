package xtcp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/misc/ip"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/connection"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/sync/workerpool"
	"go.uber.org/zap"
)

// Handler for unknown service handler.
// type Handler func(ctx context.Context, service, method string, data []byte) error
type Handler func(ctx context.Context, service, method string, data []byte, interceptor description.UnaryServerInterceptor) (interface{}, error)

type Options struct {
	WorkerSize     int `ini-name:"workerSize" long:"tcp-worker-size" description:"tcp worker size"` // numbers of worker go-routines
	BufferSize     int `ini-name:"bufferSize" long:"tcp-buffer-size" description:"tcp buffer size"` // size of buffered channel
	MaxConnections int `ini-name:"maxConnections" long:"tcp-max-connections" description:"tcp max connections"`
	Port           int

	tlsCfg                *tls.Config
	unaryInt              description.UnaryServerInterceptor
	chainUnaryInts        []description.UnaryServerInterceptor
	onconnect             func(connection.Conn)
	onclose               func(connection.Conn)
	unknownServiceHandler Handler
	// streamInt             StreamServerInterceptor
	// chainStreamInts       []StreamServerInterceptor
	// ip                    string

}

var defaultOptions = Options{
	BufferSize:     256,
	WorkerSize:     10000,
	MaxConnections: 1000,
	Port:           6000,
}

// Option sets server options.
type Option func(*Options)

// TLSCredsOption returns a Option that will set TLS credentials for server
// connections.
func TLSCredsOption(config *tls.Config) Option {
	return func(o *Options) {
		o.tlsCfg = config
	}
}

// WorkerSizeOption returns a Option that will set the number of go-routines
// in WorkerPool.
func WorkerSizeOption(workerSz int) Option {
	return func(o *Options) {
		o.WorkerSize = workerSz
	}
}

// BufferSizeOption returns a Option that is the size of buffered channel,
// for example an indicator of BufferSize256 means a size of 256.
func BufferSizeOption(indicator int) Option {
	return func(o *Options) {
		o.BufferSize = indicator
	}
}

// MaxConnections .
func MaxConnections(count int) Option {
	return func(o *Options) {
		o.MaxConnections = count
	}
}

// ConnectHandler returns a Option that will set callback to call when new
// client connected.
func ConnectHandler(cb func(connection.Conn)) Option {
	return func(o *Options) {
		o.onconnect = cb
	}
}

// CloseHandler returns a Option that will set callback to call when client
// closed.
func CloseHandler(cb func(connection.Conn)) Option {
	return func(o *Options) {
		o.onclose = cb
	}
}

// UnaryInterceptor returns a Option that sets the UnaryServerInterceptor for the
// server. Only one unary interceptor can be installed. The construction of multiple
// interceptors (e.g., chaining) can be implemented at the caller.
func UnaryInterceptor(i description.UnaryServerInterceptor) Option {
	return func(o *Options) {
		if o.unaryInt != nil {
			panic("The unary server interceptor was already set and may not be reset.")
		}
		o.unaryInt = i
	}
}

// ChainUnaryInterceptor returns a Option that specifies the chained interceptor
// for unary RPCs. The first interceptor will be the outer most,
// while the last interceptor will be the inner most wrapper around the real call.
// All unary interceptors added by this method will be chained.
func ChainUnaryInterceptor(interceptors ...description.UnaryServerInterceptor) Option {
	return func(o *Options) {
		o.chainUnaryInts = append(o.chainUnaryInts, interceptors...)
	}
}

// UnknownServiceHandler returns a Option that allows for adding a custom
// unknown service handler. The provided method is a bidi-streaming RPC service
// handler that will be invoked instead of returning the "unimplemented" gRPC
// error whenever a request is received for an unregistered service or method.
// The handling function and stream interceptor (if set) have full access to
// the ServerStream, including its Context.
func UnknownServiceHandler(handler Handler) Option {
	return func(o *Options) {
		o.unknownServiceHandler = handler
	}
}

// // IP .
// func IP(ip string) Option {
// 	return func(o *options) {
// 		o.ip = ip
// 	}
// }

// Port .
func Port(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// Server  is a server to serve TCP requests.
type Server struct {
	opts     Options
	mu       sync.Mutex // guards following
	wg       sync.WaitGroup
	conns    map[*ServerConn]bool
	services map[string]*description.ServiceInfo // service name -> service info
	lis      map[net.Listener]bool
	wps      *workerpool.WorkerPool
	// ctx      context.Context
	// cancel   context.CancelFunc
	// onconnect             func(connection.Conn)
	// onclose               func(connection.Conn)
	// unknownServiceHandler Handler
}

// New returns a new TCP server which has not started
// to serve requests yet.
func New(opt ...Option) (description.Server, func()) {
	var opts = defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	s := &Server{
		opts:     opts,
		wg:       sync.WaitGroup{},
		wps:      workerpool.New(opts.WorkerSize),
		lis:      make(map[net.Listener]bool),
		conns:    make(map[*ServerConn]bool),
		services: make(map[string]*description.ServiceInfo),
	}
	chainUnaryServerInterceptors(s)
	return s, func() {
		log.Info("xtcp is closing...")
		s.Stop()
		log.Info("xtcp is closed.")
	}
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

var _ description.ServiceRegistrar = (*Server)(nil)

// RegisterService .
func (s *Server) RegisterService(sd *description.ServiceDesc, ss interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := description.Register(&s.services, sd, ss)
	if err != nil {
		log.Fatalw("xnats register service error", "err", err)
	}
}

// Register .
func (s *Server) Register(ss interface{}, sds ...*description.ServiceDesc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sd := range sds {
		err := description.Register(&s.services, sd, ss)
		if err != nil {
			log.Fatalw("xnats register service error", "err", err)
		}
	}
}

// Serve .
func (s *Server) Serve() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.opts.Port))
	if err != nil {
		return err
	}
	s.mu.Lock()
	if s.lis == nil { // mines server is closed
		s.mu.Unlock()
		l.Close()
		return errors.New("xtcp: server has been closed")
	}
	if s.lis[l] {
		return errors.New("xtcp: listener already exist")
	}
	s.lis[l] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		if s.lis != nil && s.lis[l] {
			l.Close()
			delete(s.lis, l)
		}
		s.mu.Unlock()
	}()

	log.Infof("server start, net %s addr %s", l.Addr().Network(), l.Addr().String())

	var tempDelay time.Duration
	for {
		raw, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay >= max {
					tempDelay = max
				}
				log.Warnf("accept error %v, retrying in %d", err, tempDelay)
				<-time.After(tempDelay)
				continue
			}
			if s.lis == nil {
				log.Debugs("xtcp server is closed 1")
				return nil
			}
			log.Errors("xtcp server is closed 2", zap.Error(err))
			return err
		}
		tempDelay = 0

		if len(s.conns) >= s.opts.MaxConnections {
			log.Warnf("max connections size %d, refuse", len(s.conns))
			raw.Close()
			continue
		}

		if s.opts.tlsCfg != nil {
			raw = tls.Server(raw, s.opts.tlsCfg)
		}

		sc := newServerConn(connection.GenID(), s, raw.(*net.TCPConn))

		s.wg.Add(1)
		go func() {
			s.addConn(sc)
			s.serveConn(sc)
			s.removeConn(sc)
			s.wg.Done()
		}()
	}
}

func (s *Server) addConn(c *ServerConn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conns == nil { // means servers closed
		c.Close()
		return false
	}
	s.conns[c] = true
	return true
}

func (s *Server) removeConn(c *ServerConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conns != nil {
		delete(s.conns, c)
	}
}

func (s *Server) serveConn(sc *ServerConn) {
	if err := sc.handshake(); err != nil {
		return
	}
	// on connect
	if cb := sc.server.opts.onconnect; cb != nil {
		cb(sc)
	}
	sc.start()
	if cb := sc.server.opts.onclose; cb != nil {
		cb(sc)
	}
}

func (s *Server) Naming(nm naming.Naming) error {
	services := make(map[string]struct{})
	for name := range s.services {
		services[strings.Split(name, ".")[0]] = struct{}{}
	}
	for service := range services {
		ins := &naming.Service{
			Name:     "tcp." + service,
			Protocol: naming.TCP,
			IP:       ip.Internal(),
			Port:     s.opts.Port,
			Tag:      []string{"tcp"},
		}
		if err := nm.Register(ins); err != nil {
			return err
		}
	}
	return nil
}

// Stop .
func (s *Server) Stop() {
	s.mu.Lock()
	if s.lis == nil {
		s.mu.Unlock()
		return
	}
	listeners := s.lis
	s.lis = nil
	s.mu.Unlock()

	for l := range listeners {
		l.Close()
		log.Infof("stop accepting at address %s", l.Addr().String())
	}

	s.mu.Lock()
	for c := range s.conns {
		c.Close()
	}
	s.mu.Unlock()

	s.wg.Wait()      // wait all conns close
	s.wps.StopWait() // wait all jobs done
}
