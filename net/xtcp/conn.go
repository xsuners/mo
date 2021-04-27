package xtcp

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/connection"
	"github.com/xsuners/mo/net/message"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// // Conn is the interface that groups Write and Close methods.
// type Conn interface {
// 	RemoteAddr() net.Addr
// 	LocalAddr() net.Addr
// 	Write([]byte) error
// 	Close()
// 	ID() int64
// }

var _ connection.Conn = (*ServerConn)(nil)

// ServerConn represents a server connection to a TCP server, it implments Conn.
type ServerConn struct {
	id      int64
	server  *Server
	rawConn net.Conn
	wg      sync.WaitGroup
	sendCh  chan []byte
	ctx     context.Context
	cancel  context.CancelFunc
	closed  bool
}

func newServerConn(id int64, s *Server, c net.Conn) *ServerConn {
	sc := &ServerConn{
		id:      id,
		server:  s,
		rawConn: c,
		wg:      sync.WaitGroup{},
		sendCh:  make(chan []byte, s.opts.bufferSize),
	}
	sc.ctx, sc.cancel = context.WithCancel(context.Background())
	return sc
}

func (sc *ServerConn) start() {
	log.Infof("conn start, <%v -> %v>", sc.rawConn.LocalAddr(), sc.rawConn.RemoteAddr())

	sc.wg.Add(1)
	go func() {
		sc.readLoop()
		sc.wg.Done()
		sc.Close()
	}()

	sc.wg.Add(1)
	go func() {
		sc.writeLoop()
		sc.wg.Done()
		sc.Close()
	}()

	sc.wg.Wait()
	sc.clean()
}

// Close .
func (sc *ServerConn) Close() {
	sc.closed = true
	sc.cancel()
}

func (sc *ServerConn) clean() {
	log.Infof("conn close start ... <%v -> %v>", sc.rawConn.LocalAddr(), sc.rawConn.RemoteAddr())
	defer log.Infof("conn close done . <%v -> %v>", sc.rawConn.LocalAddr(), sc.rawConn.RemoteAddr())
	if tc, ok := sc.rawConn.(*net.TCPConn); ok {
		tc.SetLinger(0)
	}
	sc.rawConn.Close()
	close(sc.sendCh)
}

// ID .
func (sc *ServerConn) ID() int64 {
	return sc.id
}

// Write writes a message to the client.
func (sc *ServerConn) Write(message []byte) error {
	if sc.closed {
		return errors.New("conn is closed")
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, int32(len(message)))
	buf.Write(message)
	select {
	case sc.sendCh <- buf.Bytes():
		return nil
	default:
		return errors.New("xtcp: would block")
	}
}

// RemoteAddr returns the remote address of server connection.
func (sc *ServerConn) RemoteAddr() net.Addr {
	return sc.rawConn.RemoteAddr()
}

// LocalAddr returns the local address of server connection.
func (sc *ServerConn) LocalAddr() net.Addr {
	return sc.rawConn.LocalAddr()
}

func (sc *ServerConn) readLoop() {
	defer func() {
		if p := recover(); p != nil {
			log.Errorf("panics: %v", p)
		}
		log.Debug("readLoop go-routine exited")
	}()

	for {
		select {
		case <-sc.ctx.Done(): // connection closed
			log.Debug("receiving cancel signal from conn")
			return
		default:
			msg, err := message.Decode(sc.rawConn)
			if err != nil {
				if err == io.EOF {
					log.Infow("xtcp: conn closed by client side")
					return
				}
				log.Errorf("error decoding message %v", err)
				// TODO rethink
				// continue
				return
			}
			sc.process(msg)
		}
	}
}

/* writeLoop() receive message from channel, serialize it into bytes,
then blocking write into connection */
func (sc *ServerConn) writeLoop() {
	defer func() {
		if p := recover(); p != nil {
			log.Errorf("panics: %v", p)
		}
		// drain all pending messages before exit
	OuterFor:
		for {
			select {
			case pkt := <-sc.sendCh:
				if pkt != nil {
					if _, err := sc.rawConn.Write(pkt); err != nil {
						log.Errorf("error writing data %v", err)
					}
				}
			default:
				break OuterFor
			}
		}
		log.Debug("writeLoop go-routine exited")
	}()
	for {
		select {
		case <-sc.ctx.Done(): // connection closed
			log.Debug("receiving cancel signal from conn")
			return
		case pkt := <-sc.sendCh:
			if pkt != nil {
				if _, err := sc.rawConn.Write(pkt); err != nil {
					log.Errorf("error writing data %v", err)
					continue
				}
			}
		}
	}
}

// process handle a message
//
// TODO optimize with sync.Pool
func (sc *ServerConn) process(msg *message.Message) {
	ctx := context.Background()
	ctx = connection.NewContxet(ctx, sc)
	nmd := message.DecodeMetadata(msg.Metas)
	ctx = metadata.NewIncomingContext(ctx, nmd)
	srv, known := sc.server.services[msg.Service]
	if !known {
		log.Infosc(ctx, "xtcp: service not found error", zap.String("service", msg.Service))
		if handler := sc.server.opts.unknownServiceHandler; handler != nil {
			job := func() {
				out, err := handler(ctx, msg.Service, msg.Method, msg.Data, sc.server.opts.unaryInt)
				sc.response(ctx, msg, out, err)
			}
			sc.server.wps.Submit(job)
		}
		return
	}
	md, ok := srv.Method(msg.Method)
	if !ok {
		log.Errorwc(ctx, "xtcp: method not found error", "method", msg.Method)
		return
	}
	df := func(v interface{}) error {
		in, ok := v.(proto.Message)
		if !ok {
			return fmt.Errorf("xtcp: in type %T is not proto.Message", v)
		}
		return proto.Unmarshal(msg.Data, in)
	}
	job := func() {
		out, err := md.Handler(srv.Service(), ctx, df, sc.server.opts.unaryInt)
		sc.response(ctx, msg, out, err)
	}
	sc.server.wps.Submit(job)
}

// response write out message to the client.
func (sc *ServerConn) response(ctx context.Context, msg *message.Message, out interface{}, err error) {
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
				msg.Data, err = proto.Marshal(out.(proto.Message))
				if err != nil {
					log.Errorsc(ctx, "xtcp: xtcp: marshal response message error", zap.Error(err))
					return
				}
			}
		}
		// data, err := message.Encode(msg)
		data, err := proto.Marshal(msg)
		if err != nil {
			log.Errorsc(ctx, "xtcp: encode response message error", zap.Error(err))
			return
		}
		if err = sc.Write(data); err != nil {
			log.Errorsc(ctx, "xtcp: write response message error", zap.Error(err))
		}
	}
}
