package metadata

import (
	"context"
	"fmt"

	"github.com/xsuners/mo/log"
	"go.uber.org/zap"
)

var _ log.Extractor = Extractor{}

// Extractor .
type Extractor struct{}

// WithKVs .
func (Extractor) WithKVs(ctx context.Context, keysAndValues []interface{}) []interface{} {
	if ctx == nil {
		return keysAndValues
	}
	m := FromContext(ctx)
	mdKVs := []interface{}{
		"appid", m.Appid,
		"id", m.Id,
		"device", m.Device,
		"addr", m.Addr,
		"hash", m.Hash,
		"time", m.Time,
		"sn", m.Sn,
	}
	return append(mdKVs, keysAndValues...)
}

// WithArgs .
func (Extractor) WithArgs(ctx context.Context, args []interface{}) []interface{} {
	if ctx == nil {
		return args
	}
	m := FromContext(ctx)
	mdArgs := []interface{}{
		"appid" + ":", m.Appid,
		" " + "id" + ":", m.Id,
		" " + "device" + ":", m.Device,
		" " + "addr" + ":", m.Addr,
		" " + "time" + ":", m.Time,
		" " + "hash" + ":", m.Hash,
		" " + "sn" + ":", m.Sn,
		" ",
	}
	return append(mdArgs, args...)
}

// WithFormat .
func (Extractor) WithFormat(ctx context.Context, format string) string {
	if ctx == nil {
		return format
	}
	m := FromContext(ctx)
	return format + fmt.Sprintf(" appid: %d id: %d time: %d device: %d addr: %s hash: %d sn: %d", m.Appid, m.Id, m.Device, m.Addr, m.Time, m.Hash, m.Sn)
}

// WithFields .
func (Extractor) WithFields(ctx context.Context, fields []zap.Field) []zap.Field {
	if ctx == nil {
		return fields
	}
	m := FromContext(ctx)
	mdFields := []zap.Field{
		zap.Int64("appid", m.Appid),
		zap.Int64("id", m.Id),
		zap.Int32("device", m.Device),
		zap.String("addr", m.Addr),
		zap.Int64("time", m.Time),
		zap.Int64("hash", m.Hash),
		zap.Int64("sn", m.Sn),
	}
	return append(mdFields, fields...)
}
