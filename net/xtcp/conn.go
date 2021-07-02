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
	"time"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/connection"
	"github.com/xsuners/mo/net/message"
	"github.com/xsuners/mo/time/timer"
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
	id       int64
	user     connection.User
	server   *Server
	raw      *net.TCPConn
	wg       sync.WaitGroup
	mc       chan []byte
	closed   bool
	mcclosed bool
	timerid  int64
	updateAt time.Time
}

func newServerConn(id int64, s *Server, c *net.TCPConn) *ServerConn {
	return &ServerConn{
		id:       id,
		server:   s,
		raw:      c,
		wg:       sync.WaitGroup{},
		mc:       make(chan []byte, s.opts.bufferSize),
		updateAt: time.Now(),
	}
}

func (sc *ServerConn) start() {
	log.Infof("conn start, <%v -> %v>", sc.raw.LocalAddr(), sc.raw.RemoteAddr())

	sc.wg.Add(1)
	go func() {
		sc.readLoop()
		sc.wg.Done()
	}()

	sc.wg.Add(1)
	go func() {
		sc.writeLoop()
		sc.wg.Done()
	}()

	sc.check() // heartbeat check

	sc.wg.Wait()
	log.Infos("xtcp conn closed done")
}

// check .
func (sc *ServerConn) check() {
	sc.timerid = timer.RunEvery(time.Minute, func(tid int64) {
		if sc.closed {
			// 理论不会及此
			timer.Cancel(tid)
			log.Infos("理论不会及此")
			return
		}
		if time.Since(sc.updateAt) > time.Minute {
			sc.Close()
			log.Infos("xtcp heartbeat timeout")
		}
	})
}

// ID .
func (sc *ServerConn) ID() int64 {
	return sc.id
}

// User .
func (sc *ServerConn) User() connection.User {
	return sc.user
}

// Heartbeat .
func (sc *ServerConn) Heartbeat(ctx context.Context) (err error) {
	sc.updateAt = time.Now()
	return
}

// Auth .
func (sc *ServerConn) Auth(ctx context.Context, user connection.User) (err error) {
	if user == nil {
		err = errors.New("user is nil")
		return
	}
	if sc.user != nil {
		err = errors.New("authed already")
		return
	}
	sc.user = user
	return
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
	case sc.mc <- buf.Bytes():
		return nil
	default:
		return errors.New("xtcp: would block")
	}
}

// WriteMessage .
func (sc *ServerConn) WriteMessage(message proto.Message) (err error) {
	data, err := proto.Marshal(message)
	if err != nil {
		return
	}
	return sc.Write(data)
}

// RemoteAddr returns the remote address of server connection.
func (sc *ServerConn) RemoteAddr() net.Addr {
	return sc.raw.RemoteAddr()
}

// LocalAddr returns the local address of server connection.
func (sc *ServerConn) LocalAddr() net.Addr {
	return sc.raw.LocalAddr()
}

// Close .
func (sc *ServerConn) Close() {
	if sc.closed {
		return
	}
	sc.closed = true
	if sc.user != nil {
		sc.user.Disconnected()
	}
	timer.Cancel(sc.timerid)
	if len(sc.mc) < 1 {
		log.Debugs("xtcp mc closed 1")
		sc.mcclosed = true
		close(sc.mc)
	}
	if err := sc.raw.CloseRead(); err != nil {
		log.Errors("xtcp conn close read error", zap.Error(err))
	}
}

// readLoop .
func (sc *ServerConn) readLoop() {
	defer sc.Close()
	for {
		msg, err := message.Decode(sc.raw)
		if err != nil {
			if err == io.EOF {
				log.Infos("xtcp read loop closed 1")
				return
			}
			if sc.closed {
				log.Infos("xtcp read loop closed 2")
				return
			}
			log.Errors("xtcp read loop closed 3", zap.Error(err))
			return
		}
		sc.process(msg)
	}
}

// writeLoop .
func (sc *ServerConn) writeLoop() {
	defer func() {
		if !sc.mcclosed {
			close(sc.mc)
			sc.mcclosed = true
			log.Infos("xtcp mc closed 2")
		}
		if err := sc.raw.CloseWrite(); err != nil {
			log.Infos("xtcp conn close write error:", zap.Error(err))
			return
		}
	}()
	for {
		m, ok := <-sc.mc
		if !ok {
			log.Infos("xtcp write loop closed 1")
			return
		}
		if _, err := sc.raw.Write(m); err != nil {
			log.Errorf("xtcp error writing data %v", err)
			continue
		}
		if sc.closed {
			if len(sc.mc) < 1 {
				log.Infos("xtcp write loop closed 2")
				return
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
			sc.wg.Add(1)
			job := func() {
				sc.wg.Done()
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
	sc.wg.Add(1)
	job := func() {
		sc.wg.Done()
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
