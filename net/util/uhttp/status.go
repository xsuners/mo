package uhttp

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/xsuners/mo/net/message"
	"github.com/xsuners/mo/net/util/meta"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// PrepareHeaders .
func PrepareHeaders(ctx context.Context, hdr http.Header) (context.Context, error) {
	out, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		out = metadata.MD{}
	}
	if auth := hdr.Get("Authorization"); auth != "" {
		out.Set("authorization", auth)
	}
	if mt := hdr.Get(meta.HK); mt != "" {
		out.Set(meta.MK, mt)
		bmt, err := base64.StdEncoding.DecodeString(mt)
		if err != nil {
			return ctx, err
		}
		smt := &message.Metadata{}
		if err := proto.Unmarshal(bmt, smt); err != nil {
			return ctx, err
		}
		ctx = meta.NewContext(ctx, smt)
	}
	for key, val := range hdr {
		if strings.HasPrefix(strings.ToLower(key), "x-goog-") {
			out.Set(key, val...)
		}
	}
	return metadata.NewOutgoingContext(ctx, out), nil
}

// Code2Status .
func Code2Status(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	}
	return http.StatusInternalServerError
}
