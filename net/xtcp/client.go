package xtcp

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/connection"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/time/timer"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var _ connection.Conn = (*ClientConn)(nil)

// Codec .
type Codec interface {
	Decode(raw io.Reader) ([]byte, error)
	Encode(data []byte) ([]byte, error)
}

// dialOptions configure a Dial call. dialOptions are set by the DialOption
// values passed to Dial.
type dialOptions struct {
	onconnect      func(connection.Conn)
	onclose        func(connection.Conn)
	onmessage      func(data []byte) error
	codec          Codec
	unaryInt       description.UnaryClientInterceptor
	chainUnaryInts []description.UnaryClientInterceptor
	// streamInt StreamClientInterceptor
	// chainStreamInts []StreamClientInterceptor
	// nopts []nats.Option
	bufferSize int // size of buffered channel
}

// DialOption configures how we set up the connection.
type DialOption interface {
	apply(*dialOptions)
}

// EmptyDialOption does not alter the dial configuration. It can be embedded in
// another structure to build custom dial options.
//
// Experimental
//
// Notice: This type is EXPERIMENTAL and may be changed or removed in a
// later release.
type EmptyDialOption struct{}

func (EmptyDialOption) apply(*dialOptions) {}

// funcDialOption wraps a function that modifies dialOptions into an
// implementation of the DialOption interface.
type funcDialOption struct {
	f func(*dialOptions)
}

func (fdo *funcDialOption) apply(do *dialOptions) {
	fdo.f(do)
}

func newFuncDialOption(f func(*dialOptions)) *funcDialOption {
	return &funcDialOption{
		f: f,
	}
}

// WithUnaryInterceptor returns a DialOption that specifies the interceptor for
// unary RPCs.
func WithUnaryInterceptor(f description.UnaryClientInterceptor) DialOption {
	return newFuncDialOption(func(o *dialOptions) {
		o.unaryInt = f
	})
}

// WithChainUnaryInterceptor returns a DialOption that specifies the chained
// interceptor for unary RPCs. The first interceptor will be the outer most,
// while the last interceptor will be the inner most wrapper around the real call.
// All interceptors added by this method will be chained, and the interceptor
// defined by WithUnaryInterceptor will always be prepended to the chain.
func WithChainUnaryInterceptor(interceptors ...description.UnaryClientInterceptor) DialOption {
	return newFuncDialOption(func(o *dialOptions) {
		o.chainUnaryInts = append(o.chainUnaryInts, interceptors...)
	})
}

// ConnectOption returns a DialOption that will set callback to call when new
// client connected.
func ConnectOption(cb func(connection.Conn)) DialOption {
	return newFuncDialOption(func(o *dialOptions) {
		o.onconnect = cb
	})
}

// CloseOption returns a DialOption that will set callback to call when client
// closed.
func CloseOption(cb func(connection.Conn)) DialOption {
	return newFuncDialOption(func(o *dialOptions) {
		o.onclose = cb
	})
}

// MessageOption returns a DialOption that will set callback to call when new
// client connected.
func MessageOption(cb func([]byte) error) DialOption {
	return newFuncDialOption(func(o *dialOptions) {
		o.onmessage = cb
	})
}

// CustomCodec returns a DialOption that will set callback to call when client
// closed.
func CustomCodec(codec Codec) DialOption {
	return newFuncDialOption(func(o *dialOptions) {
		o.codec = codec
	})
}

// // WithNatsOption config under nats .
// func WithNatsOption(opt nats.Option) DialOption {
// 	return newFuncDialOption(func(o *dialOptions) {
// 		o.nopts = append(o.nopts, opt)
// 	})
// }

func defaultDialOptions() dialOptions {
	return dialOptions{
		codec: Codecc{},
		// disableRetry:    !envconfig.Retry,
		// healthCheckFunc: internal.HealthCheckFunc,
		// copts: transport.ConnectOptions{
		// 	WriteBufferSize: defaultWriteBufSize,
		// 	ReadBufferSize:  defaultReadBufSize,
		// },
		// resolveNowBackoff: internalbackoff.DefaultExponential.Backoff,
		// withProxy:         true,
		bufferSize: 1024,
	}
}

// ClientConn .
type ClientConn struct {
	id       int64
	user     connection.User
	addr     string
	opts     dialOptions
	raw      net.Conn
	wg       sync.WaitGroup
	sendCh   chan []byte
	timerid  int64
	updateAt time.Time
	closed   bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewClientConn returns a new client connection which has not started to
// serve requests yet.
func NewClientConn(netid int64, c net.Conn, opt ...DialOption) *ClientConn {
	var opts = defaultDialOptions()
	for _, o := range opt {
		o.apply(&opts)
	}
	// TODO intercepter
	cc := &ClientConn{
		id:       netid,
		addr:     c.RemoteAddr().String(),
		opts:     opts,
		raw:      c,
		wg:       sync.WaitGroup{},
		sendCh:   make(chan []byte, opts.bufferSize),
		updateAt: time.Now(),
	}
	cc.ctx, cc.cancel = context.WithCancel(context.Background())
	return cc
}

// Start .
func (cc *ClientConn) Start() {
	log.Infof("conn start, <%v -> %v>", cc.raw.LocalAddr(), cc.raw.RemoteAddr())

	if onconnect := cc.opts.onconnect; onconnect != nil {
		onconnect(cc)
	}

	cc.wg.Add(1)
	go func() {
		cc.readLoop()
		cc.wg.Done()
		cc.cancel()
	}()

	cc.wg.Add(1)
	go func() {
		cc.writeLoop()
		cc.wg.Done()
		cc.cancel()
	}()

	cc.check()
	cc.wg.Wait()

	// callback on close
	if onclose := cc.opts.onclose; onclose != nil {
		onclose(cc)
	}
}

// check .
func (cc *ClientConn) check() {
	cc.timerid = timer.RunEvery(time.Minute, func(tid int64) {
		if cc.closed {
			// 理论不会及此
			timer.Cancel(tid)
			log.Infos("理论不会及此")
			return
		}
		if time.Since(cc.updateAt) > time.Minute {
			cc.Close()
			log.Infos("heartbeat timeout")
		}
	})
}

// Close .
func (cc *ClientConn) Close() {
	if cc.closed {
		return
	}
	log.Infow("xtcp: conn close start", "loacl", cc.raw.LocalAddr(), "remote", cc.raw.RemoteAddr())
	defer log.Infow("xtcp: conn close done")
	cc.closed = true
	timer.Cancel(cc.timerid)
	cc.cancel()
	cc.raw.Close()
	close(cc.sendCh)
}

// Write writes a message to the client.
func (cc *ClientConn) Write(message []byte) (err error) {
	select {
	case cc.sendCh <- message:
		return nil
	default:
		return errors.New("xtcp: would block")
	}
}

// WriteMessage .
func (cc *ClientConn) WriteMessage(message proto.Message) (err error) {
	data, err := proto.Marshal(message)
	if err != nil {
		return
	}
	return cc.Write(data)
}

// ID .
func (cc *ClientConn) ID() int64 {
	return cc.id
}

// User .
func (cc *ClientConn) User() connection.User {
	return cc.user
}

// Heartbeat .
func (cc *ClientConn) Heartbeat(ctx context.Context) (err error) {
	cc.updateAt = time.Now()
	return
}

// Auth .
func (cc *ClientConn) Auth(ctx context.Context, user connection.User) (err error) {
	if user == nil {
		err = errors.New("user is nil")
		return
	}
	if cc.user != nil {
		err = errors.New("authed already")
		return
	}
	cc.user = user
	return nil
}

// RemoteAddr .
func (cc *ClientConn) RemoteAddr() net.Addr {
	return cc.raw.RemoteAddr()
}

// LocalAddr .
func (cc *ClientConn) LocalAddr() net.Addr {
	return cc.raw.LocalAddr()
}

func (cc *ClientConn) readLoop() {
	defer func() {
		if p := recover(); p != nil {
			log.Errorf("panics: %v", p)
		}
		log.Debug("xtcp: read loop exited")
	}()

	for {
		data, err := cc.opts.codec.Decode(cc.raw)
		if err != nil {
			if err == io.EOF {
				log.Infow("xtcp: client conn closed by server side")
				return
			}
			if cc.closed {
				log.Infos("conn closed")
				return
			}
			log.Errorw("xtcp: client decoding message error", "err", err)
			return
		}
		// TODO
		// log.Infow("xtcp: client revced message", "message", data)
		if cc.opts.onmessage != nil {
			if err = cc.opts.onmessage(data); err != nil {
				log.Debugs("om message error", zap.Error(err))
			}
		}
		select {
		case <-cc.ctx.Done(): // connection closed
			return
		default:
		}
	}
}

func (cc *ClientConn) writeLoop() {
	defer func() {
		if p := recover(); p != nil {
			log.Errorf("panics: %v", p)
		}
		// drain all pending messages before exit
	OuterFor:
		for {
			select {
			case pkt := <-cc.sendCh:
				if pkt != nil {
					if _, err := cc.raw.Write(pkt); err != nil {
						log.Errorf("error writing data %v", err)
					}
				}
			default:
				break OuterFor
			}
		}
		log.Debug("xtcp: write loop exited")
	}()

	for {
		select {
		case <-cc.ctx.Done(): // connection closed
			return
		case pkt := <-cc.sendCh:
			if pkt != nil {
				data, err := cc.opts.codec.Encode(pkt)
				if err != nil {
					log.Errors("encode data error", zap.Error(err))
					continue
				}
				if _, err := cc.raw.Write(data); err != nil {
					log.Errors("writing data error", zap.Error(err))
					continue
				}
			}
		}
	}
}
