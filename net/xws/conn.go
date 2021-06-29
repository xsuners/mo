package xws

import (
	"context"
	"errors"
	"io"
	"net"

	"github.com/gobwas/ws"
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/connection"
	"github.com/xsuners/mo/net/message"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// // Conn is used in options.
// type Conn interface {
// 	RemoteAddr() net.Addr
// 	LocalAddr() net.Addr
// 	Write(message []byte) error
// 	Close()
// }

var _ connection.Conn = (*wrappedConn)(nil)

type wrappedConn struct {
	id     int64
	user   connection.User
	raw    net.Conn
	server *Server
	closed bool
	// ctx    context.Context
	// cancel context.CancelFunc
}

func newWrappedConn(id int64, s *Server, c net.Conn) *wrappedConn {
	wc := &wrappedConn{
		id:     id,
		raw:    c,
		server: s,
	}
	// ctx := context.Background()
	// wc.ctx, wc.cancel = context.WithCancel(ctx)
	return wc
}

func (wc *wrappedConn) Close() {
	// wc.cancel()
	if wc.closed {
		return
	}
	if wc.user != nil {
		wc.user.Disconnected()
	}
	wc.closed = true
	wc.raw.Close()
}

func (wc *wrappedConn) Write(message []byte) error {
	if wc.closed {
		return errors.New("xws: conn is closed")
	}
	header := ws.Header{
		Fin:    true,
		OpCode: ws.OpBinary,
		Length: int64(len(message)),
	}
	err := ws.WriteHeader(wc.raw, header)
	if err != nil {
		return err
	}
	_, err = wc.raw.Write(message)
	return err
	// return wsutil.WriteMessage(wc.raw, ws.StateServerSide, ws.OpBinary, message)
}

// WriteMessage .
func (wc *wrappedConn) WriteMessage(message proto.Message) (err error) {
	data, err := proto.Marshal(message)
	if err != nil {
		return
	}
	return wc.Write(data)
}

// func (wc *wrappedConn) Drain() {
// 	log.Infow("xws: todo drain conn")
// }

// ID .
func (wc *wrappedConn) ID() int64 {
	return wc.id
}

// ID .
func (sc *wrappedConn) User() connection.User {
	return sc.user
}

// Heartbeat .
func (sc *wrappedConn) Heartbeat(ctx context.Context) (err error) {
	// TODO
	return
}

// Auth .
func (sc *wrappedConn) Auth(ctx context.Context, user connection.User) (err error) {
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

// RemoteAddr returns the remote address of server connection.
func (wc *wrappedConn) RemoteAddr() net.Addr {
	return wc.raw.RemoteAddr()
}

// LocalAddr returns the local address of server connection.
func (wc *wrappedConn) LocalAddr() net.Addr {
	return wc.raw.LocalAddr()
}

func (wc *wrappedConn) Serve(handle func(ctx context.Context, msg *message.Message)) {
	for {
		header, err := ws.ReadHeader(wc.raw)
		if err != nil {
			log.Errorw("xws: read header error, continue", "err", err)
			return
			// continue
		}

		// if wc.closed {
		// 	log.Infos("xws: conn closed by server")
		// 	return
		// }

		// select {
		// case <-wc.ctx.Done():
		// 	log.Infow("xws: conn closed on server side")
		// 	return
		// default:
		// }

		payload := make([]byte, header.Length)
		_, err = io.ReadFull(wc.raw, payload)
		if err != nil {
			log.Errorw("xws: read payload error, continue", "err", err)
			continue
		}
		if header.Masked {
			ws.Cipher(payload, header.Mask, 0)
		}

		// Reset the Masked flag, server frames must not be masked as
		// RFC6455 says.
		// header.Masked = false
		// if err := ws.WriteHeader(conn, header); err != nil {
		// 	// handle error
		// 	panic(err)
		// }

		// log.Debug("opcode", header.OpCode)

		if header.OpCode == ws.OpClose {
			log.Info("xws: opcode is close, close the conn")
			return
		}

		msg := new(message.Message)
		if err = wc.server.opts.codec.Unmarshal(payload, msg); err != nil {
			log.Errorw("xws: decode payload error", "err", err)
			return
			// continue
		}

		// TODO sync.Pool
		ctx := context.Background()
		ctx = connection.NewContxet(ctx, wc)
		// nmd := metadata.New(msg.Metadata)
		nmd := message.DecodeMetadata(msg.Metas)
		ctx = metadata.NewIncomingContext(ctx, nmd)

		handle(ctx, msg)
	}
}
