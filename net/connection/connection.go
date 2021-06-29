package connection

import (
	"context"
	"net"
)

// Conn is used in options.
type Conn interface {
	ID() int64
	Close()
	Write(message []byte) error
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	Heartbeat(ctx context.Context) (err error)
	Auth(ctx context.Context, user User) (err error)
}

type User interface {
	Disconnected()
}

type connectionKey struct{}

// FromContext .
func FromContext(ctx context.Context) (conn Conn, ok bool) {
	conn, ok = ctx.Value(connectionKey{}).(Conn)
	return
}

// NewContxet .
func NewContxet(ctx context.Context, conn Conn) context.Context {
	return context.WithValue(ctx, connectionKey{}, conn)
}
