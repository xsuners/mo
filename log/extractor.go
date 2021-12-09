package log

import (
	"context"

	"go.uber.org/zap"
)

// Extractor .
type Extractor interface {
	WithKVs(ctx context.Context, keysAndValues []interface{}) []interface{}
	WithArgs(ctx context.Context, args []interface{}) []interface{}
	WithFormat(ctx context.Context, format string) string
	WithFields(ctx context.Context, fields []zap.Field) []zap.Field
}
