package connection

import (
	"context"
	"net"
)

// Conn is used in options.
type Conn interface {
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	Write(message []byte) error
	Close()
	ID() int64
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
