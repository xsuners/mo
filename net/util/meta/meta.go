package meta

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/xsuners/mo/net/encoding"
	jsonc "github.com/xsuners/mo/net/encoding/json"
	protoc "github.com/xsuners/mo/net/encoding/proto"
	"github.com/xsuners/mo/net/message"
	"google.golang.org/grpc/metadata"
)

////////////////////////////

var (
	// MK name is used in grpc metadata.
	MK  = proto.MessageName((*message.Metadata)(nil))
	JMK = MK + ".json"
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
	if len(mds) > 0 {
		data, _ := base64.StdEncoding.DecodeString(mds[0])
		smd := &message.Metadata{}
		err := proto.Unmarshal(data, smd)
		if err != nil {
			return nil, false
		}
		return smd, true
	}
	mds = md.Get(JMK)
	if len(mds) > 0 {
		data, _ := base64.StdEncoding.DecodeString(mds[0])
		smd := &message.Metadata{}
		err := json.Unmarshal(data, smd)
		if err != nil {
			return nil, false
		}
		return smd, true
	}
	return nil, false
}

// Base64 .
// TODO 对 codec 的支持更自由
func Base64(mt *message.Metadata, codec encoding.Codec) (*message.Meta, error) {
	data, err := codec.Marshal(mt)
	if err != nil {
		return nil, err
	}
	switch codec.Name() {
	case jsonc.Name:
		return &message.Meta{
			Name:  JMK,
			Value: base64.StdEncoding.EncodeToString(data),
		}, nil
	case protoc.Name:
		return &message.Meta{
			Name:  MK,
			Value: base64.StdEncoding.EncodeToString(data),
		}, nil
	}
	return nil, errors.New("unknown codec")
}
