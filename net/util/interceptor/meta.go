package interceptor

import (
	"context"

	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/message"
	"github.com/xsuners/mo/net/util/meta"
)

// MetaServerInterceptor .
func MetaServerInterceptor() description.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *description.UnaryServerInfo, handler description.UnaryHandler) (interface{}, error) {
		m, ok := meta.FromIncomingContext(ctx)
		if !ok {
			m = &message.Metadata{}
		}
		ctx = meta.NewContext(ctx, m)
		return handler(ctx, req)
	}
}

// MetaClientInterceptor .
func MetaClientInterceptor() description.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc description.UnaryClient, invoker description.UnaryInvoker, opts ...description.CallOption) error {
		m := meta.FromContext(ctx)
		ctx = meta.NewOutgoingContext(ctx, m)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
