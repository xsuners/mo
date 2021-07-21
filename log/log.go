package log

import (
	"context"

	"github.com/xsuners/mo/log/extractor"
	"github.com/xsuners/mo/net/util/ip"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

// // Config .
// type Config struct {
// 	Path  string `json:"path"`
// 	Level Level  `json:"level"`
// 	Tags  []Tag  `json:"tags"`
// }

type option struct {
	extractors []extractor.Extractor
	zopts      []zap.Option
	path       string
	level      Level
	tags       []Tag
}

func defaultOptions() option {
	return option{
		level: LevelDebug,
		path:  "/dev/null",
	}
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
func WithExtractor(exts ...extractor.Extractor) Option {
	return newFuncOption(func(o *option) {
		o.extractors = append(o.extractors, exts...)
	})
}

// LogLevel .
func LogLevel(l Level) Option {
	return newFuncOption(func(o *option) {
		o.level = l
	})
}

// Path .
func Path(path string) Option {
	return newFuncOption(func(o *option) {
		o.path = path
	})
}

// Tags .
func Tags(tags ...Tag) Option {
	return newFuncOption(func(o *option) {
		o.tags = append(o.tags, tags...)
	})
}

type Log struct {
	opt    option
	logger *zap.Logger
	suger  *zap.SugaredLogger
}

// // Logger .
// type Logger struct {
// 	origin *zap.Logger
// 	opt    option
// }

// // Sugar .
// type Sugar struct {
// 	origin *zap.SugaredLogger
// 	opt    option
// }

var log *Log

// var logger *Logger
// var suger *Sugar

// New .
func New(opts ...Option) (*Log, func()) {
	opt := defaultOptions()
	for _, o := range opts {
		o.apply(&opt)
	}

	log = new(Log)
	log.opt = opt

	// core := zapcore.NewCore(
	// 	zapcore.NewJSONEncoder(zapcore.EncoderConfig{
	// 		TimeKey:        "ts",
	// 		LevelKey:       "level",
	// 		NameKey:        "logger",
	// 		CallerKey:      "caller",
	// 		MessageKey:     "msg",
	// 		StacktraceKey:  "stacktrace",
	// 		LineEnding:     zapcore.DefaultLineEnding,
	// 		EncodeLevel:    zapcore.LowercaseLevelEncoder,
	// 		EncodeTime:     zapcore.ISO8601TimeEncoder,
	// 		EncodeDuration: zapcore.SecondsDurationEncoder,
	// 		EncodeCaller:   zapcore.ShortCallerEncoder,
	// 	}),
	// 	zapcore.AddSync(&lumberjack.Logger{
	// 		Filename:   opt.path,
	// 		MaxSize:    100, // megabytes
	// 		MaxBackups: 3,
	// 		MaxAge:     28, // days
	// 	}),
	// 	zapLevel(opt.level),
	// )

	// opt.zopts = append(opt.zopts, zapFields(opt.tags))
	// opt.zopts = append(opt.zopts, zap.AddCaller())
	// opt.zopts = append(opt.zopts, zap.AddCallerSkip(1))

	// log.logger = zap.New(core, opt.zopts...)
	// log.suger = log.logger.Sugar()

	log.logger, _ = zap.NewProduction(zap.AddCallerSkip(1), zapFields(opt.tags))
	log.suger = log.logger.Sugar()

	return log, func() {
		Infos("log is closing...")
		log.Close()
		Infos("log is closed.")
	}
}

// Close close log
func (log *Log) Close() {
	if log.logger != nil {
		log.logger.Sync()
	}
	if log.suger != nil {
		log.suger.Sync()
	}
}

// Close close log
func Close() {
	if log.logger != nil {
		log.logger.Sync()
	}
	if log.suger != nil {
		log.suger.Sync()
	}
}

// func zapLevel(lvl Level) zapcore.LevelEnabler {
// 	switch lvl {
// 	case LevelDebug:
// 		return zap.DebugLevel
// 	case LevelInfo:
// 		return zap.InfoLevel
// 	case LevelWarn:
// 		return zap.WarnLevel
// 	case LevelError:
// 		return zap.ErrorLevel
// 	case LevelFatal:
// 		return zap.FatalLevel
// 	}
// 	return zap.InfoLevel
// }

func zapFields(tags []Tag) zap.Option {
	fs := []zap.Field{}
	for _, tag := range tags {
		fs = append(fs, zap.Field{
			Key:    tag.Key,
			Type:   zapcore.StringType,
			String: tag.Value,
		})
	}
	fs = append(fs, zapcore.Field{
		Key:    "host",
		Type:   zapcore.StringType,
		String: ip.Internal(),
	})
	return zap.Fields(fs...)
}

// Debug .
func Debug(args ...interface{}) {
	log.suger.Debug(args...)
}

// Debugf .
func Debugf(format string, args ...interface{}) {
	log.suger.Debugf(format, args...)
}

// Debugw .
func Debugw(msg string, keysAndValues ...interface{}) {
	log.suger.Debugw(msg, keysAndValues...)
}

// Debugs .
func Debugs(msg string, fields ...zap.Field) {
	log.logger.Debug(msg, fields...)
}

// Info .
func Info(args ...interface{}) {
	log.suger.Info(args...)
}

// Infof .
func Infof(format string, args ...interface{}) {
	log.suger.Infof(format, args...)
}

// Infow .
func Infow(msg string, keysAndValues ...interface{}) {
	log.suger.Infow(msg, keysAndValues...)
}

// Infos .
func Infos(msg string, fields ...zap.Field) {
	log.logger.Info(msg, fields...)
}

// Warn .
func Warn(args ...interface{}) {
	log.suger.Warn(args...)
}

// Warnf .
func Warnf(format string, args ...interface{}) {
	log.suger.Warnf(format, args...)
}

// Warnw .
func Warnw(msg string, keysAndValues ...interface{}) {
	log.suger.Warnw(msg, keysAndValues...)
}

// Warns .
func Warns(msg string, fields ...zap.Field) {
	log.logger.Warn(msg, fields...)
}

// Error .
func Error(args ...interface{}) {
	log.suger.Error(args...)
}

// Errorf .
func Errorf(format string, args ...interface{}) {
	log.suger.Errorf(format, args...)
}

// Errorw .
func Errorw(msg string, keysAndValues ...interface{}) {
	log.suger.Errorw(msg, keysAndValues...)
}

// Errors .
func Errors(msg string, fields ...zap.Field) {
	log.logger.Error(msg, fields...)
}

// Panic .
func Panic(args ...interface{}) {
	log.suger.Panic(args...)
}

// Panicf .
func Panicf(format string, args ...interface{}) {
	log.suger.Panicf(format, args...)
}

// Panicw .
func Panicw(msg string, keysAndValues ...interface{}) {
	log.suger.Panicw(msg, keysAndValues...)
}

// Panics .
func Panics(msg string, fields ...zap.Field) {
	log.logger.Panic(msg, fields...)
}

// Fatal .
func Fatal(args ...interface{}) {
	log.suger.Fatal(args...)
}

// Fatalf .
func Fatalf(format string, args ...interface{}) {
	log.suger.Fatalf(format, args...)
}

// Fatalw .
func Fatalw(msg string, keysAndValues ...interface{}) {
	log.suger.Fatalw(msg, keysAndValues...)
}

// Fatals .
func Fatals(msg string, fields ...zap.Field) {
	log.logger.Fatal(msg, fields...)
}

// Debugc .
func Debugc(ctx context.Context, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		args = ex.WithMDArgs(ctx, args)
	}
	log.suger.Debug(args...)
}

// Debugfc .
func Debugfc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		format = ex.WithMDFormat(ctx, format)
	}
	log.suger.Debugf(format, args...)
}

// Debugwc .
func Debugwc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range log.opt.extractors {
		keysAndValues = ex.WithMDKVs(ctx, keysAndValues)
	}
	log.suger.Debugw(msg, keysAndValues...)
}

// Debugsc .
func Debugsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range log.opt.extractors {
		fields = ex.WithMDFields(ctx, fields)
	}
	log.logger.Debug(msg, fields...)
}

// Infoc .
func Infoc(ctx context.Context, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		args = ex.WithMDArgs(ctx, args)
	}
	log.suger.Info(args...)
}

// Infofc .
func Infofc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		format = ex.WithMDFormat(ctx, format)
	}
	log.suger.Infof(format, args...)
}

// Infowc .
func Infowc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range log.opt.extractors {
		keysAndValues = ex.WithMDKVs(ctx, keysAndValues)
	}
	log.suger.Infow(msg, keysAndValues...)
}

// Infosc .
func Infosc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range log.opt.extractors {
		fields = ex.WithMDFields(ctx, fields)
	}
	log.logger.Info(msg, fields...)
}

// Warnc .
func Warnc(ctx context.Context, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		args = ex.WithMDArgs(ctx, args)
	}
	log.suger.Warn(args...)
}

// Warnfc .
func Warnfc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		format = ex.WithMDFormat(ctx, format)
	}
	log.suger.Warnf(format, args...)
}

// Warnwc .
func Warnwc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range log.opt.extractors {
		keysAndValues = ex.WithMDKVs(ctx, keysAndValues)
	}
	log.suger.Warnw(msg, keysAndValues...)
}

// Warnsc .
func Warnsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range log.opt.extractors {
		fields = ex.WithMDFields(ctx, fields)
	}
	log.logger.Warn(msg, fields...)
}

// Errorc .
func Errorc(ctx context.Context, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		args = ex.WithMDArgs(ctx, args)
	}
	log.suger.Error(args...)
}

// Errorfc .
func Errorfc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		format = ex.WithMDFormat(ctx, format)
	}
	log.suger.Errorf(format, args...)
}

// Errorwc .
func Errorwc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range log.opt.extractors {
		keysAndValues = ex.WithMDKVs(ctx, keysAndValues)
	}
	log.suger.Errorw(msg, keysAndValues...)
}

// Errorsc .
func Errorsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range log.opt.extractors {
		fields = ex.WithMDFields(ctx, fields)
	}
	log.logger.Error(msg, fields...)
}

// Panicc .
func Panicc(ctx context.Context, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		args = ex.WithMDArgs(ctx, args)
	}
	log.suger.Panic(args...)
}

// Panicfc .
func Panicfc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		format = ex.WithMDFormat(ctx, format)
	}
	log.suger.Panicf(format, args...)
}

// Panicwc .
func Panicwc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range log.opt.extractors {
		keysAndValues = ex.WithMDKVs(ctx, keysAndValues)
	}
	log.suger.Panicw(msg, keysAndValues...)
}

// Panicsc .
func Panicsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range log.opt.extractors {
		fields = ex.WithMDFields(ctx, fields)
	}
	log.logger.Panic(msg, fields...)
}

// Fatalc .
func Fatalc(ctx context.Context, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		args = ex.WithMDArgs(ctx, args)
	}
	log.suger.Fatal(args...)
}

// Fatalfc .
func Fatalfc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		format = ex.WithMDFormat(ctx, format)
	}
	log.suger.Fatalf(format, args...)
}

// Fatalwc .
func Fatalwc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range log.opt.extractors {
		keysAndValues = ex.WithMDKVs(ctx, keysAndValues)
	}
	log.suger.Fatalw(msg, keysAndValues...)
}

// Fatalsc .
func Fatalsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range log.opt.extractors {
		fields = ex.WithMDFields(ctx, fields)
	}
	log.logger.Fatal(msg, fields...)
}

func init() {
	origin, err := zap.NewDevelopment(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
	log = &Log{
		logger: origin,
		suger:  origin.Sugar(),
		opt:    defaultOptions(),
	}
}
