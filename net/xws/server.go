package xws

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobwas/ws"
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/misc/ip"
	"github.com/xsuners/mo/misc/xrand"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/connection"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/encoding"
	"github.com/xsuners/mo/net/encoding/json"
	"github.com/xsuners/mo/net/encoding/proto"
	"github.com/xsuners/mo/net/message"
	"github.com/xsuners/mo/sync/event"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

type serverWorkerData struct {
	ctx  context.Context
	wg   *sync.WaitGroup
	conn *wrappedConn
	data *message.Message
}

// Server is a gRPC server to serve RPC requests.
type Server struct {
	opts Options

	mu    sync.Mutex // guards following
	lis   map[net.Listener]bool
	conns map[*wrappedConn]bool
	serve bool
	// drain    bool
	services map[string]*description.ServiceInfo // service name -> service info
	// events   trace.EventLog

	cv      *sync.Cond     // signaled when connections close for GracefulStop
	serveWG sync.WaitGroup // counts active Serve goroutines for GracefulStop

	quit *event.Event
	done *event.Event
	// channelzRemoveOnce sync.Once

	// channelzID int64 // channelz unique identification number
	// czData     *channelzData

	serverWorkerChannels []chan *serverWorkerData

	// ctx    context.Context
	// cancel context.CancelFunc
}

// Handler .
type Handler func(ctx context.Context, service, method string, data []byte, interceptor description.UnaryServerInterceptor) (interface{}, error)

type Options struct {
	NumServerWorkers uint32 `ini-name:"numServerWorkers" long:"ws-workers" description:"ws server workers number"`
	Port             int

	// creds                 credentials.TransportCredentials
	// codec          Codec
	connectHandler func(connection.Conn)
	closeHandler   func(connection.Conn)
	// cp                    Compressor
	// dc                    Decompressor
	// unaryInt              UnaryServerInterceptor
	unaryInt       description.UnaryServerInterceptor
	chainUnaryInts []description.UnaryServerInterceptor
	// chainStreamInts       []StreamServerInterceptor
	// inTapHandle           tap.ServerInHandle
	// statsHandler          stats.Handler
	// maxConcurrentStreams  uint32
	// maxReceiveMessageSize int
	// maxSendMessageSize    int
	// unknownStreamDesc     *StreamDesc
	// keepaliveParams       keepalive.ServerParameters
	// keepalivePolicy       keepalive.EnforcementPolicy
	// initialWindowSize     int32
	// initialConnWindowSize int32
	// writeBufferSize       int
	// readBufferSize        int
	connectionTimeout time.Duration
	// maxHeaderListSize     *uint32
	// headerTableSize       *uint32
	unknownServiceHandler Handler
}

var defaultOptions = Options{
	// maxReceiveMessageSize: defaultServerMaxReceiveMessageSize,
	// maxSendMessageSize:    defaultServerMaxSendMessageSize,
	connectionTimeout: 120 * time.Second,
	NumServerWorkers:  100,
	Port:              5000,
	// codec:             NewBaseCodec(),
	// writeBufferSize:       defaultWriteBufSize,
	// readBufferSize:        defaultReadBufSize,
}

// A Option sets options such as credentials, codec and keepalive parameters, etc.
type Option interface {
	apply(*Options)
}

// EmptyOption does not alter the server configuration. It can be embedded
// in another structure to build custom server options.
//
// # Experimental
//
// Notice: This type is EXPERIMENTAL and may be changed or removed in a
// later release.
type EmptyOption struct{}

func (EmptyOption) apply(*Options) {}

// funcOption wraps a function that modifies options into an
// implementation of the Option interface.
type funcOption struct {
	f func(*Options)
}

func (fdo *funcOption) apply(do *Options) {
	fdo.f(do)
}

func newFuncOption(f func(*Options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// ConnectionTimeout returns a Option that sets the timeout for
// connection establishment (up to and including HTTP/2 handshaking) for all
// new connections.  If this is not set, the default is 120 seconds.  A zero or
// negative value will result in an immediate timeout.
//
// # Experimental
//
// Notice: This API is EXPERIMENTAL and may be changed or removed in a
// later release.
func ConnectionTimeout(d time.Duration) Option {
	return newFuncOption(func(o *Options) {
		o.connectionTimeout = d
	})
}

// NumStreamWorkers returns a Option that sets the number of worker
// goroutines that should be used to process incoming streams. Setting this to
// zero (default) will disable workers and spawn a new goroutine for each
// stream.
//
// # Experimental
//
// Notice: This API is EXPERIMENTAL and may be changed or removed in a
// later release.
func NumStreamWorkers(numServerWorkers uint32) Option {
	// TODO: If/when this API gets stabilized (i.e. stream workers become the
	// only way streams are processed), change the behavior of the zero value to
	// a sane default. Preliminary experiments suggest that a value equal to the
	// number of CPUs available is most performant; requires thorough testing.
	return newFuncOption(func(o *Options) {
		o.NumServerWorkers = numServerWorkers
	})
}

// ConnectHandler .
func ConnectHandler(f func(connection.Conn)) Option {
	return newFuncOption(func(o *Options) {
		o.connectHandler = f
	})
}

// CloseHandler .
func CloseHandler(f func(connection.Conn)) Option {
	return newFuncOption(func(o *Options) {
		o.closeHandler = f
	})
}

// UnaryInterceptor returns a Option that sets the UnaryServerInterceptor for the
// server. Only one unary interceptor can be installed. The construction of multiple
// interceptors (e.g., chaining) can be implemented at the caller.
func UnaryInterceptor(i description.UnaryServerInterceptor) Option {
	return newFuncOption(func(o *Options) {
		if o.unaryInt != nil {
			panic("The unary server interceptor was already set and may not be reset.")
		}
		o.unaryInt = i
	})
}

// ChainUnaryInterceptor returns a Option that specifies the chained interceptor
// for unary RPCs. The first interceptor will be the outer most,
// while the last interceptor will be the inner most wrapper around the real call.
// All unary interceptors added by this method will be chained.
func ChainUnaryInterceptor(interceptors ...description.UnaryServerInterceptor) Option {
	return newFuncOption(func(o *Options) {
		o.chainUnaryInts = append(o.chainUnaryInts, interceptors...)
	})
}

// UnknownServiceHandler returns a Option that allows for adding a custom
// unknown service handler. The provided method is a bidi-streaming RPC service
// handler that will be invoked instead of returning the "unimplemented" gRPC
// error whenever a request is received for an unregistered service or method.
// The handling function and stream interceptor (if set) have full access to
// the ServerStream, including its Context.
func UnknownServiceHandler(handler Handler) Option {
	return newFuncOption(func(o *Options) {
		o.unknownServiceHandler = handler
	})
}

func Port(port int) Option {
	return newFuncOption(func(o *Options) {
		o.Port = port
	})
}

// serverWorkerResetThreshold defines how often the stack must be reset. Every
// N requests, by spawning a new goroutine in its place, a worker can reset its
// stack so that large stacks don't live in memory forever. 2^16 should allow
// each goroutine stack to live for at least a few seconds in a typical
// workload (assuming a QPS of a few thousand requests/sec).
const serverWorkerResetThreshold = 1 << 16

// serverWorkers blocks on a *transport.Stream channel forever and waits for
// data to be fed by serveStreams. This allows different requests to be
// processed by the same goroutine, removing the need for expensive stack
// re-allocations (see the runtime.morestack problem [1]).
//
// [1] https://github.com/golang/go/issues/18138
func (s *Server) serverWorker(ch chan *serverWorkerData) {
	// To make sure all server workers don't reset at the same time, choose a
	// random number of iterations before resetting.
	threshold := serverWorkerResetThreshold + xrand.Intn(serverWorkerResetThreshold)
	for completed := 0; completed < threshold; completed++ {
		data, ok := <-ch
		if !ok {
			return
		}
		s.process(data.ctx, data.conn, data.data)
		data.wg.Done()
	}
	go s.serverWorker(ch)
}

// initServerWorkers creates worker goroutines and channels to process incoming
// connections to reduce the time spent overall on runtime.morestack.
func (s *Server) initServerWorkers() {
	s.serverWorkerChannels = make([]chan *serverWorkerData, s.opts.NumServerWorkers)
	for i := uint32(0); i < s.opts.NumServerWorkers; i++ {
		s.serverWorkerChannels[i] = make(chan *serverWorkerData)
		go s.serverWorker(s.serverWorkerChannels[i])
	}
}

func (s *Server) stopServerWorkers() {
	for i := uint32(0); i < s.opts.NumServerWorkers; i++ {
		close(s.serverWorkerChannels[i])
	}
}

// New creates a gRPC server which has no service registered and has not
// started to accept requests yet.
func New(opt ...Option) (description.Server, func()) {
	opts := defaultOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	s := &Server{
		lis:      make(map[net.Listener]bool),
		opts:     opts,
		conns:    make(map[*wrappedConn]bool),
		services: make(map[string]*description.ServiceInfo),
		quit:     event.NewEvent(),
		done:     event.NewEvent(),
		// czData:   new(channelzData),
	}

	// TODO
	chainUnaryServerInterceptors(s)
	// chainStreamServerInterceptors(s)

	s.cv = sync.NewCond(&s.mu)

	// if EnableTracing {
	// 	_, file, line, _ := runtime.Caller(1)
	// 	s.events = trace.NewEventLog("grpc.Server", fmt.Sprintf("%s:%d", file, line))
	// }

	if s.opts.NumServerWorkers > 0 {
		s.initServerWorkers()
	}

	// l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.opts.ip, s.opts.port))
	// if err != nil {
	// 	return
	// }

	// s.Serve(l)
	return s, func() {
		log.Info("xws is closing...")
		s.Stop()
		log.Info("xws is closed.")
	}
	// ctx := context.Background()
	// s.ctx, s.cancel = context.WithCancel(ctx)
	// if channelz.IsOn() {
	// 	s.channelzID = channelz.RegisterServer(&channelzServer{s}, "")
	// }
	// return
}

var _ description.ServiceRegistrar = (*Server)(nil)

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

// RegisterService .
func (s *Server) RegisterService(sd *description.ServiceDesc, ss interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.serve {
		log.Fatalf("xws: Server.RegisterService after Server.Serve for %s", sd.ServiceName)
	}
	err := description.Register(&s.services, sd, ss)
	if err != nil {
		log.Fatalw("xws: register service error", "err", err)
	}
}

// Register .
func (s *Server) Register(ss interface{}, sds ...*description.ServiceDesc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.serve {
		log.Fatal("xws: Server.RegisterService after Server.Serve")
	}
	for _, sd := range sds {
		err := description.Register(&s.services, sd, ss)
		if err != nil {
			log.Fatalw("xws: register service error", "err", err)
		}
	}
}

// Serve accepts incoming connections on the listener lis, creating a new
// ServerTransport and service goroutine for each. The service goroutines
// read gRPC requests and then call the registered handlers to reply to them.
// Serve returns when lis.Accept fails with fatal errors.  lis will be closed when
// this method returns.
// Serve will return a non-nil error unless Stop or GracefulStop is called.
func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.opts.Port))
	if err != nil {
		return err
	}
	s.mu.Lock()
	// s.printf("serving")
	s.serve = true
	if s.lis == nil {
		// Serve called after Stop or GracefulStop.
		s.mu.Unlock()
		lis.Close()
		return errors.New("xws: the server has been stopped")
	}

	s.serveWG.Add(1)
	defer func() {
		s.serveWG.Done()
		if s.quit.HasFired() {
			// Stop or GracefulStop called; block until done and return nil.
			<-s.done.Done()
		}
	}()

	// ls := &listenSocket{Listener: lis}
	s.lis[lis] = true

	// if channelz.IsOn() {
	// 	ls.channelzID = channelz.RegisterListenSocket(ls, s.channelzID, lis.Addr().String())
	// }
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		if s.lis != nil && s.lis[lis] {
			lis.Close()
			delete(s.lis, lis)
		}
		s.mu.Unlock()
	}()

	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		rawConn, err := lis.Accept()
		if err != nil {
			if ne, ok := err.(interface {
				Temporary() bool
			}); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				// s.mu.Lock()
				// // s.printf("Accept error: %v; retrying in %v", err, tempDelay)
				// s.mu.Unlock()
				timer := time.NewTimer(tempDelay)
				select {
				case <-timer.C:
				// case <-s.ctx.Done():
				// 	timer.Stop()
				// 	return nil
				case <-s.quit.Done():
					timer.Stop()
					return nil
				}
				continue
			}
			// s.mu.Lock()
			// // s.printf("done serving; Accept = %v", err)
			// s.mu.Unlock()
			if s.quit.HasFired() {
				return nil
			}
			return err
		}
		tempDelay = 0
		// Start a new goroutine to deal with rawConn so we don't stall this Accept
		// loop goroutine.
		//
		// Make sure we account for the goroutine so GracefulStop doesn't nil out
		// s.conns before this conn can be added.
		s.serveWG.Add(1)
		go func() {
			s.handleConn(rawConn)
			s.serveWG.Done()
		}()
	}
}

// handleConn forks a goroutine to handle a just-accepted connection that
// has not had any I/O performed on it yet.
func (s *Server) handleConn(conn net.Conn) {
	if s.quit.HasFired() {
		conn.Close()
		return
	}

	wc := newWrappedConn(connection.GenID(), s, conn)

	u := ws.Upgrader{
		OnHeader: func(key, value []byte) (err error) {
			log.Infof("xws: non-websocket header: %q=%q", key, value)
			return
		},
		ProtocolCustom: func(b []byte) (string, bool) {
			ok, err := regexp.Match("proto", b)
			if err != nil {
				log.Errors("xws: protocol error", zap.Error(err))
				return "", false
			}
			if ok {
				wc.codec = encoding.GetCodec(proto.Name)
				return "protobuf", true
			}
			ok, err = regexp.Match("json", b)
			if err != nil {
				log.Errors("xws: protocol error", zap.Error(err))
				return "", false
			}
			if ok {
				wc.codec = encoding.GetCodec(json.Name)
				return "json", true
			}
			return "", false
		},
	}
	_, err := u.Upgrade(conn)
	if err != nil {
		if err == io.EOF {
			log.Infos("check")
		} else {
			log.Errors("xws: upgrade error", zap.Error(err))
		}
		conn.Close()
		return
	}

	if !s.addConn(wc) {
		return
	}
	go func() {
		s.serveStreams(wc)
		s.removeConn(wc)
	}()
}

func (s *Server) addConn(conn *wrappedConn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conns == nil {
		conn.raw.Close()
		return false
	}
	// if s.drain {
	// 	// Transport added after we drained our existing conns: drain it
	// 	// immediately.
	// 	conn.Drain()
	// }
	s.conns[conn] = true
	return true
}

func (s *Server) removeConn(conn *wrappedConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conns != nil {
		delete(s.conns, conn)
		s.cv.Broadcast()
	}
}

func (s *Server) serveStreams(conn *wrappedConn) {
	// must serve and process all over then to close
	defer conn.Close()
	// TODO 精细化关闭
	// defer func() {
	// // make client side gracefal close
	// header := ws.Header{Fin: true, OpCode: ws.OpClose}
	// if err := ws.WriteHeader(conn.raw, header); err != nil {
	// 	log.Errorw("xws: write close message error", "err", err)
	// }

	// if err := conn.raw.Close(); err != nil {
	// 	log.Errorw("xws: close conn error", "err", err)
	// }
	// }()

	var wg sync.WaitGroup
	var roundRobinCounter uint32

	// on conn created
	if cb := s.opts.connectHandler; cb != nil {
		cb(conn)
	}

	conn.Serve(func(ctx context.Context, msg *message.Message) {
		wg.Add(1)
		if s.opts.NumServerWorkers < 1 {
			go func() {
				defer wg.Done()
				s.process(ctx, conn, msg)
			}()
			return
		}

		data := &serverWorkerData{ctx: ctx, conn: conn, wg: &wg, data: msg}

		select {
		case s.serverWorkerChannels[atomic.AddUint32(&roundRobinCounter, 1)%s.opts.NumServerWorkers] <- data:
		default: // If all msg workers are busy, fallback to the default code path.
			go func() {
				s.process(ctx, conn, msg)
				wg.Done()
			}()
		}
	})

	// on conn close
	if cb := s.opts.closeHandler; cb != nil {
		cb(conn)
	}

	wg.Wait()
}

func (s *Server) process(ctx context.Context, conn *wrappedConn, msg *message.Message) {
	srv, known := s.services[msg.Service]
	if !known {
		// desc := fmt.Sprintf("xws: get service (%s) error", msg.Service)
		// reply(ctx, conn, 1, desc, nil)
		log.Infosc(ctx, "xws: get service error", zap.String("service", msg.Service))
		if handler := s.opts.unknownServiceHandler; handler != nil {
			out, err := handler(ctx, msg.Service, msg.Method, msg.Data, s.opts.unaryInt)
			response(ctx, conn, msg, out, err)
		}
		return
	}

	md, ok := srv.Method(msg.Method)
	if !ok {
		// desc := fmt.Sprintf("xws: get method (%s) error", msg.Method)
		// reply(ctx, conn, 1, desc, nil)
		log.Errorsc(ctx, "xws: get method error", zap.String("method", msg.Method))
		return
	}

	df := func(v interface{}) error {
		// req, ok := v.(proto.Message)
		// if !ok {
		// 	return fmt.Errorf("in type %T is not proto.Message", v)
		// }
		return conn.codec.Unmarshal(msg.Data, v)
	}

	out, err := md.Handler(srv.Service(), ctx, df, s.opts.unaryInt)
	if err != nil {
		// reply(ctx, conn, 1, err.Error(), nil)
		log.Warnsc(ctx, "xws: handle message error", zap.Error(err))
		return
	}
	response(ctx, conn, msg, out, err)

	// om, ok := out.(proto.Message)
	// if !ok {
	// 	desc := fmt.Sprintf("xws: out message (%T) not proto.Message", out)
	// 	reply(ctx, conn, 1, desc, nil)
	// 	return
	// }

	// data, err := proto.Marshal(om)
	// if err != nil {
	// 	desc := "xws: marshal out message error"
	// 	reply(ctx, conn, 1, desc, nil)
	// 	return
	// }

	// reply(ctx, conn, 0, "", data)
}

// response write out message to the client.
func response(ctx context.Context, conn *wrappedConn, msg *message.Message, out interface{}, err error) {
	if len(msg.Messageid) > 0 { // 需要返回
		if err != nil {
			st := status.Convert(err)
			msg.Code = st.Proto().Code
			msg.Desc = st.Proto().Message
			log.Debugsc(ctx, "xtcp: call method get error", zap.Error(err))
		} else {
			if frame, ok := out.(*message.Frame); ok { // proxy respons id frame
				msg.Data = frame.Data
			} else {
				msg.Data, err = conn.codec.Marshal(out)
				if err != nil {
					log.Errorsc(ctx, "xtcp: xtcp: marshal response message error", zap.Error(err))
					return
				}
			}
		}
		data, err := conn.codec.Marshal(msg)
		if err != nil {
			log.Errorsc(ctx, "xtcp: encode response message error", zap.Error(err))
			return
		}
		if err = conn.Write(data); err != nil {
			log.Errorsc(ctx, "xtcp: write response message error", zap.Error(err))
		}
	}
}

// func reply(ctx context.Context, conn *wrappedConn, code int32, desc string, data []byte) {
// 	response := &message.WSResponse{}
// 	if code != 0 {
// 		response.Code = code
// 		response.Desc = desc
// 		log.Errorc(ctx, desc)
// 	} else if data == nil {
// 		response.Code = code
// 		response.Desc = "xws: internal error, code id 0 but data is nil"
// 		log.Errorc(ctx, desc)
// 	} else {
// 		response.Data = data
// 	}

// 	data, err := proto.Marshal(response)
// 	if err != nil {
// 		log.Errorwc(ctx, "xws: marshal response error", "err", err)
// 		return
// 	}

// 	if err = conn.WriteMessage(data); err != nil {
// 		log.Errorwc(ctx, "xws: response error", "err", err)
// 	}
// }

func (s *Server) Naming(nm naming.Naming) error {
	for name := range s.services {
		ins := &naming.Service{
			Name:     name,
			Protocol: naming.WS,
			IP:       ip.Internal(),
			Port:     s.opts.Port,
			Tag:      []string{"ws"},
		}
		if err := nm.Register(ins); err != nil {
			return err
		}
	}
	return nil
}

// Stop stops the gRPC server gracefully. It stops the server from
// accepting new connections and RPCs and blocks until all the pending RPCs are
// finished.
func (s *Server) Stop() {
	// s.cancel()
	s.quit.Fire()
	defer s.done.Fire()

	// s.channelzRemoveOnce.Do(func() {
	// 	if channelz.IsOn() {
	// 		channelz.RemoveEntry(s.channelzID)
	// 	}
	// })
	s.mu.Lock()
	if s.conns == nil {
		s.mu.Unlock()
		return
	}

	for lis := range s.lis {
		lis.Close()
	}
	s.lis = nil
	// if !s.drain {
	// 	for conn := range s.conns {
	// 		conn.Drain()
	// 	}
	// 	s.drain = true
	// }

	for conn := range s.conns {
		conn.Close()
	}

	if s.opts.NumServerWorkers > 0 {
		s.stopServerWorkers()
	}

	// Wait for serving threads to be ready to exit.  Only then can we be sure no
	// new conns will be created.
	s.mu.Unlock()
	s.serveWG.Wait()
	s.mu.Lock()

	for len(s.conns) != 0 {
		s.cv.Wait()
	}
	s.conns = nil
	// if s.events != nil {
	// 	s.events.Finish()
	// 	s.events = nil
	// }
	s.mu.Unlock()
}
