package uhttp

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/xsuners/mo/metadata"
	"github.com/xsuners/mo/net/encoding"
	"google.golang.org/grpc/codes"
	md "google.golang.org/grpc/metadata"
)

// PrepareHeaders .
func PrepareHeaders(ctx context.Context, h http.Header, cc encoding.Codec) (context.Context, error) {
	out, ok := md.FromOutgoingContext(ctx)
	if !ok {
		out = md.MD{}
	}
	if auth := h.Get("Authorization"); auth != "" {
		out.Set("authorization", auth)
	}
	if mt := h.Get(metadata.HK); mt != "" {
		out.Set(metadata.MK, mt)
		bmt, err := base64.StdEncoding.DecodeString(mt)
		if err != nil {
			return ctx, err
		}
		smt := &metadata.Metadata{}
		if err := cc.Unmarshal(bmt, smt); err != nil {
			return ctx, err
		}
		ctx = metadata.NewContext(ctx, smt)
	}
	for key, val := range h {
		if strings.HasPrefix(strings.ToLower(key), "x-goog-") {
			out.Set(key, val...)
		}
	}
	return md.NewOutgoingContext(ctx, out), nil
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
