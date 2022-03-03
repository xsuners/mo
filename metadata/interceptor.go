package metadata

import (
	"context"

	"github.com/xsuners/mo/net/description"
)

// ServerInterceptor .
func ServerInterceptor() description.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *description.UnaryServerInfo, handler description.UnaryHandler) (interface{}, error) {
		m, ok := FromIncomingContext(ctx)
		if !ok {
			m = &Metadata{}
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
