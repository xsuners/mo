package meta

import (
	"context"
	"encoding/base64"
	"math"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/xsuners/mo/net/message"
	"google.golang.org/grpc/metadata"
)

////////////////////////////

var (
	// MK name is used in grpc metadata.
	MK = proto.MessageName((*message.Metadata)(nil))
	// HK name is used in http header
	HK = "x-mo-meta"
)

type mk struct{}

// NewContext .
func NewContext(ctx context.Context, md *message.Metadata) context.Context {
	return context.WithValue(ctx, mk{}, md)
}

// FromContext .
func FromContext(ctx context.Context) *message.Metadata {
	if md, ok := ctx.Value(mk{}).(*message.Metadata); ok {
		return md
	}
	return &message.Metadata{
		Sn: math.MaxInt64,
	}
}

// Clone .
func Clone(in *message.Metadata) (out *message.Metadata) {
	if in == nil {
		return &message.Metadata{
			Time: time.Now().Unix(),
		}
	}
	return proto.Clone(in).(*message.Metadata)
}

// NewOutgoingContext .
func NewOutgoingContext(ctx context.Context, smd *message.Metadata) context.Context {
	data, _ := proto.Marshal(smd)
	encodeString := base64.StdEncoding.EncodeToString(data)
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		md = md.Copy()
		md.Set(MK, encodeString)
	} else {
		md = metadata.New(map[string]string{
			MK: encodeString,
		})
	}
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx
}

// FromIncomingContext .
func FromIncomingContext(ctx context.Context) (*message.Metadata, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, false
	}
	mds := md.Get(MK)
	if len(mds) < 1 {
		return nil, false
	}
	decodeBytes, _ := base64.StdEncoding.DecodeString(mds[0])
	smd := &message.Metadata{}
	err := proto.Unmarshal(decodeBytes, smd)
	if err != nil {
		return nil, false
	}
	return smd, true
}

// Base64 .
func Base64(mt *message.Metadata) string {
	data, _ := proto.Marshal(mt)
	return base64.StdEncoding.EncodeToString(data)
}
