package metadata

import (
	"context"

	"github.com/xsuners/mo/misc/ip"
	"github.com/xsuners/mo/net/description"
)

type Option func(*Metadata)

func Addr() Option {
	return func(m *Metadata) {
		m.Addr = ip.Internal()
	}
}

// ServerInterceptor .
func ServerInterceptor(opts ...Option) description.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *description.UnaryServerInfo, handler description.UnaryHandler) (interface{}, error) {
		m, ok := FromIncomingContext(ctx)
		if !ok {
			m = &Metadata{
				Ints: make(map[string]int64),
				Strs: make(map[string]string),
				Objs: make(map[string][]byte),
			}
		}
		for _, opt := range opts {
			opt(m)
		}
		ctx = NewContext(ctx, m)
		return handler(ctx, req)
	}
}

// ClientInterceptor .
func ClientInterceptor() description.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc description.UnaryClient, invoker description.UnaryInvoker, opts ...description.CallOption) error {
		m := FromContext(ctx)
		ctx = NewOutgoingContext(ctx, m)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
