package publisher

import (
	"context"
	"fmt"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/message"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func (pub *Publisher) Publish(ctx context.Context, in proto.Message, opts ...description.CallOption) error {
	data, err := proto.Marshal(in)
	if err != nil {
		return err
	}

	// TODO 包一下
	co := copool.Get().(*CallOptions)
	defer copool.Put(co)

	co.Timeout = pub.dopts.defaultTimeout
	co.Subject = pub.dopts.defaultSubject
	for _, o := range opts {
		o.Apply(co)
	}

	// TODO use sync.Pool
	request := &message.Message{
		// Service: service,
		// Method:  method,
		Data: data,
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		request.Metas = message.EncodeMetadata(md)
	}

	data, err = proto.Marshal(request)
	if err != nil {
		return err
	}

	subj := string(in.ProtoReflect().Descriptor().FullName())

	if !co.WaitResponse { // pub-sub mode
		if err := pub.conn.Publish(subj, data); err != nil {
			return err
		}
		return nil
	}

	msg, err := pub.conn.Request(subj, data, co.Timeout)
	if err != nil {
		log.Errorsc(ctx, "invoke:Request", zap.String("subject", subj), zap.Error(err))
		return err
	}
	response := &message.Message{} // TODO use sync.Pool
	if err = proto.Unmarshal(msg.Data, response); err != nil {
		return err
	}
	// TODO 讲错误信息包装成status返回
	if response.Code != 0 {
		return fmt.Errorf("xnats: response (%s) error", response.Desc)
	}
	// err = proto.Unmarshal(response.Data, out)
	return nil
}
