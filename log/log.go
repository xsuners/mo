package log

import (
	"context"

	"github.com/xsuners/mo/log/extracter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Level .
type Level int8

const (
	// LevelDebug .
	LevelDebug Level = iota
	// LevelInfo .
	LevelInfo
	// LevelWarn .
	LevelWarn
	// LevelError .
	LevelError
	// LevelFatal .
	LevelFatal
)

// // Extracter .
// type Extracter interface {
// 	WithMDKVs(ctx context.Context, keysAndValues []interface{}) []interface{}
// 	WithMDArgs(ctx context.Context, args []interface{}) []interface{}
// 	WithMDFormat(ctx context.Context, format string) string
// 	WithMDFields(ctx context.Context, fields []zap.Field) []zap.Field
// }

// Tag .
type Tag struct {
	Key   string
	Value string
}

// Config .
type Config struct {
	Path  string `json:"path"`
	Level Level  `json:"level"`
	Tags  []Tag  `json:"tags"`
}

type option struct {
	extractors []extracter.Extracter
	zopts      []zap.Option
}

func defaultOptions() option {
	return option{}
}

// A Option sets options such as credentials, codec and keepalive parameters, etc.
type Option interface {
	apply(*option)
}

// EmptyOption does not alter the server configuration. It can be embedded
// in another structure to build custom server options.
//
// Experimental
//
// Notice: This type is EXPERIMENTAL and may be changed or removed in a
// later release.
type EmptyOption struct{}

func (EmptyOption) apply(*option) {}

// funcOption wraps a function that modifies option into an
// implementation of the Option interface.
type funcOption struct {
	f func(*option)
}

func (fdo *funcOption) apply(do *option) {
	fdo.f(do)
}

func newFuncOption(f func(*option)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// WithZapOption config under nats .
func WithZapOption(opts ...zap.Option) Option {
	return newFuncOption(func(o *option) {
		o.zopts = append(o.zopts, opts...)
	})
}

// WithExtractor config under nats .
func WithExtractor(exts ...extracter.Extracter) Option {
	return newFuncOption(func(o *option) {
		o.extractors = append(o.extractors, exts...)
	})
}

// Logger .
type Logger struct {
	origin *zap.Logger
	opt    option
}

// Sugar .
type Sugar struct {
	origin *zap.SugaredLogger
	opt    option
}

var logger *Logger
var suger *Sugar

// Init .
func Init(c *Config, opts ...Option) (err error) {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   c.Path,
			MaxSize:    100, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
		}),
		zapLevel(c.Level),
	)

	opt := defaultOptions()
	for _, o := range opts {
		o.apply(&opt)
	}

	opt.zopts = append(opt.zopts, zapFields(c.Tags))
	opt.zopts = append(opt.zopts, zap.AddCaller())
	opt.zopts = append(opt.zopts, zap.AddCallerSkip(1))

	logger.origin = zap.New(core, opt.zopts...)
	logger.opt = opt

	suger.origin = logger.origin.Sugar()
	suger.opt = opt

	return nil
}

// Close close log
func Close() {
	if origin := logger.origin; origin != nil {
		origin.Sync()
	}
	if origin := suger.origin; origin != nil {
		origin.Sync()
	}
}

func zapLevel(lvl Level) zapcore.LevelEnabler {
	switch lvl {
	case LevelDebug:
		return zap.DebugLevel
	case LevelInfo:
		return zap.InfoLevel
	case LevelWarn:
		return zap.WarnLevel
	case LevelError:
		return zap.ErrorLevel
	case LevelFatal:
		return zap.FatalLevel
	}
	return zap.InfoLevel
}

func zapFields(tags []Tag) zap.Option {
	fs := []zap.Field{}
	for _, tag := range tags {
		fs = append(fs, zap.Field{
			Key:    tag.Key,
			Type:   zapcore.StringType,
			String: tag.Value,
		})
	}
	return zap.Fields(fs...)
}

// Debug .
func Debug(args ...interface{}) {
	suger.origin.Debug(args...)
}

// Debugf .
func Debugf(format string, args ...interface{}) {
	suger.origin.Debugf(format, args...)
}

// Debugw .
func Debugw(msg string, keysAndValues ...interface{}) {
	suger.origin.Debugw(msg, keysAndValues...)
}

// Debugs .
func Debugs(msg string, fields ...zap.Field) {
	logger.origin.Debug(msg, fields...)
}

// Info .
func Info(args ...interface{}) {
	suger.origin.Info(args...)
}

// Infof .
func Infof(format string, args ...interface{}) {
	suger.origin.Infof(format, args...)
}

// Infow .
func Infow(msg string, keysAndValues ...interface{}) {
	suger.origin.Infow(msg, keysAndValues...)
}

// Infos .
func Infos(msg string, fields ...zap.Field) {
	logger.origin.Info(msg, fields...)
}

// Warn .
func Warn(args ...interface{}) {
	suger.origin.Warn(args...)
}

// Warnf .
func Warnf(format string, args ...interface{}) {
	suger.origin.Warnf(format, args...)
}

// Warnw .
func Warnw(msg string, keysAndValues ...interface{}) {
	suger.origin.Warnw(msg, keysAndValues...)
}

// Warns .
func Warns(msg string, fields ...zap.Field) {
	logger.origin.Warn(msg, fields...)
}

// Error .
func Error(args ...interface{}) {
	suger.origin.Error(args...)
}

// Errorf .
func Errorf(format string, args ...interface{}) {
	suger.origin.Errorf(format, args...)
}

// Errorw .
func Errorw(msg string, keysAndValues ...interface{}) {
	suger.origin.Errorw(msg, keysAndValues...)
}

// Errors .
func Errors(msg string, fields ...zap.Field) {
	logger.origin.Error(msg, fields...)
}

// Panic .
func Panic(args ...interface{}) {
	suger.origin.Panic(args...)
}

// Panicf .
func Panicf(format string, args ...interface{}) {
	suger.origin.Panicf(format, args...)
}

// Panicw .
func Panicw(msg string, keysAndValues ...interface{}) {
	suger.origin.Panicw(msg, keysAndValues...)
}

// Panics .
func Panics(msg string, fields ...zap.Field) {
	logger.origin.Panic(msg, fields...)
}

// Fatal .
func Fatal(args ...interface{}) {
	suger.origin.Fatal(args...)
}

// Fatalf .
func Fatalf(format string, args ...interface{}) {
	suger.origin.Fatalf(format, args...)
}

// Fatalw .
func Fatalw(msg string, keysAndValues ...interface{}) {
	suger.origin.Fatalw(msg, keysAndValues...)
}

// Fatals .
func Fatals(msg string, fields ...zap.Field) {
	logger.origin.Fatal(msg, fields...)
}

// Debugc .
func Debugc(ctx context.Context, args ...interface{}) {
	for _, ex := range suger.opt.extractors {
		args = ex.WithMDArgs(ctx, args)
	}
	suger.origin.Debug(args...)
}

// Debugfc .
func Debugfc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range suger.opt.extractors {
		format = ex.WithMDFormat(ctx, format)
	}
	suger.origin.Debugf(format, args...)
}

// Debugwc .
func Debugwc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range suger.opt.extractors {
		keysAndValues = ex.WithMDKVs(ctx, keysAndValues)
	}
	suger.origin.Debugw(msg, keysAndValues...)
}

// Debugsc .
func Debugsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range logger.opt.extractors {
		fields = ex.WithMDFields(ctx, fields)
	}
	logger.origin.Debug(msg, fields...)
}

// Infoc .
func Infoc(ctx context.Context, args ...interface{}) {
	for _, ex := range suger.opt.extractors {
		args = ex.WithMDArgs(ctx, args)
	}
	suger.origin.Info(args...)
}

// Infofc .
func Infofc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range suger.opt.extractors {
		format = ex.WithMDFormat(ctx, format)
	}
	suger.origin.Infof(format, args...)
}

// Infowc .
func Infowc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range suger.opt.extractors {
		keysAndValues = ex.WithMDKVs(ctx, keysAndValues)
	}
	suger.origin.Infow(msg, keysAndValues...)
}

// Infosc .
func Infosc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range logger.opt.extractors {
		fields = ex.WithMDFields(ctx, fields)
	}
	logger.origin.Info(msg, fields...)
}

// Warnc .
func Warnc(ctx context.Context, args ...interface{}) {
	for _, ex := range suger.opt.extractors {
		args = ex.WithMDArgs(ctx, args)
	}
	suger.origin.Warn(args...)
}

// Warnfc .
func Warnfc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range suger.opt.extractors {
		format = ex.WithMDFormat(ctx, format)
	}
	suger.origin.Warnf(format, args...)
}

// Warnwc .
func Warnwc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range suger.opt.extractors {
		keysAndValues = ex.WithMDKVs(ctx, keysAndValues)
	}
	suger.origin.Warnw(msg, keysAndValues...)
}

// Warnsc .
func Warnsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range logger.opt.extractors {
		fields = ex.WithMDFields(ctx, fields)
	}
	logger.origin.Warn(msg, fields...)
}

// Errorc .
func Errorc(ctx context.Context, args ...interface{}) {
	for _, ex := range suger.opt.extractors {
		args = ex.WithMDArgs(ctx, args)
	}
	suger.origin.Error(args...)
}

// Errorfc .
func Errorfc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range suger.opt.extractors {
		format = ex.WithMDFormat(ctx, format)
	}
	suger.origin.Errorf(format, args...)
}

// Errorwc .
func Errorwc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range suger.opt.extractors {
		keysAndValues = ex.WithMDKVs(ctx, keysAndValues)
	}
	suger.origin.Errorw(msg, keysAndValues...)
}

// Errorsc .
func Errorsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range logger.opt.extractors {
		fields = ex.WithMDFields(ctx, fields)
	}
	logger.origin.Error(msg, fields...)
}

// Panicc .
func Panicc(ctx context.Context, args ...interface{}) {
	for _, ex := range suger.opt.extractors {
		args = ex.WithMDArgs(ctx, args)
	}
	suger.origin.Panic(args...)
}

// Panicfc .
func Panicfc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range suger.opt.extractors {
		format = ex.WithMDFormat(ctx, format)
	}
	suger.origin.Panicf(format, args...)
}

// Panicwc .
func Panicwc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range suger.opt.extractors {
		keysAndValues = ex.WithMDKVs(ctx, keysAndValues)
	}
	suger.origin.Panicw(msg, keysAndValues...)
}

// Panicsc .
func Panicsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range logger.opt.extractors {
		fields = ex.WithMDFields(ctx, fields)
	}
	logger.origin.Panic(msg, fields...)
}

// Fatalc .
func Fatalc(ctx context.Context, args ...interface{}) {
	for _, ex := range suger.opt.extractors {
		args = ex.WithMDArgs(ctx, args)
	}
	suger.origin.Fatal(args...)
}

// Fatalfc .
func Fatalfc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range suger.opt.extractors {
		format = ex.WithMDFormat(ctx, format)
	}
	suger.origin.Fatalf(format, args...)
}

// Fatalwc .
func Fatalwc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range logger.opt.extractors {
		keysAndValues = ex.WithMDKVs(ctx, keysAndValues)
	}
	suger.origin.Fatalw(msg, keysAndValues...)
}

// Fatalsc .
func Fatalsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range logger.opt.extractors {
		fields = ex.WithMDFields(ctx, fields)
	}
	logger.origin.Fatal(msg, fields...)
}

func init() {
	origin, err := zap.NewDevelopment(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
	logger = &Logger{
		origin: origin,
	}
	suger = &Sugar{
		origin: logger.origin.Sugar(),
	}
}
