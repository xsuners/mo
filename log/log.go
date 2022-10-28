package log

import (
	"context"
	"fmt"
	"os"

	"github.com/xsuners/mo/misc/ip"
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

// Tag .
type Tag struct {
	Key   string
	Value string
}

type Options struct {
	Path    string `ini-name:"path" long:"log-path" description:"log file path"`
	Level   Level  `ini-name:"level" long:"log-level" description:"log level"`
	Console bool   `ini-name:"console" long:"log-console" description:"log to console"`

	tags       []Tag
	extractors []Extractor
	zopts      []zap.Option
	alarmer    Alarmer
}

func defaultOptions() Options {
	return Options{
		Level:   LevelDebug,
		Path:    "/dev/null",
		Console: false,
		alarmer: &empty{},
	}
}

// A Option sets options such as credentials, codec and keepalive parameters, etc.
type Option func(*Options)

// WithZapOption config under nats .
func WithZapOption(opts ...zap.Option) Option {
	return func(o *Options) {
		o.zopts = append(o.zopts, opts...)
	}
}

// WithExtractor config under nats .
func WithExtractor(exts ...Extractor) Option {
	return func(o *Options) {
		o.extractors = append(o.extractors, exts...)
	}
}

// WithAlarmer config under nats .
func WithAlarmer(a Alarmer) Option {
	return func(o *Options) {
		o.alarmer = a
	}
}

// LogLevel .
func LogLevel(l Level) Option {
	return func(o *Options) {
		o.Level = l
	}
}

// Path .
func Path(path string) Option {
	return func(o *Options) {
		o.Path = path
	}
}

// Console .
func Console() Option {
	return func(o *Options) {
		o.Console = true
	}
}

// Tags .
func Tags(tags ...Tag) Option {
	return func(o *Options) {
		o.tags = append(o.tags, tags...)
	}
}

type Log struct {
	opt    Options
	logger *zap.Logger
	suger  *zap.SugaredLogger
}

var log *Log

// New .
func New(opts ...Option) (*Log, func()) {
	log = &Log{
		opt: defaultOptions(),
	}
	for _, o := range opts {
		o(&log.opt)
	}

	var cores []zapcore.Core

	cores = append(cores, zapcore.NewCore(
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
			EncodeCaller:   zapcore.FullCallerEncoder,
		}),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   log.opt.Path,
			MaxSize:    200, // megabytes
			MaxBackups: 3,
			MaxAge:     7, // days
		}),
		zapLevel(log.opt.Level),
	))

	if log.opt.Console {
		// The bundled Config struct only supports the most common configuration
		// options. More complex needs, like splitting logs between multiple files
		// or writing to non-file outputs, require use of the zapcore package.
		//
		// In this example, imagine we're both sending our logs to Kafka and writing
		// them to the console. We'd like to encode the console output and the Kafka
		// topics differently, and we'd also like special treatment for
		// high-priority logs.

		// First, define our level-handling logic.
		highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		})
		lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl < zapcore.ErrorLevel
		})

		// Assume that we have clients for two Kafka topics. The clients implement
		// zapcore.WriteSyncer and are safe for concurrent use. (If they only
		// implement io.Writer, we can use zapcore.AddSync to add a no-op Sync
		// method. If they're not safe for concurrent use, we can add a protecting
		// mutex with zapcore.Lock.)
		// topicDebugging := zapcore.AddSync(ioutil.Discard)
		// topicErrors := zapcore.AddSync(ioutil.Discard)

		// High-priority output should also go to standard error, and low-priority
		// output should also go to standard out.
		consoleDebugging := zapcore.Lock(os.Stdout)
		consoleErrors := zapcore.Lock(os.Stderr)

		// Optimize the Kafka output for machine consumption and the console output
		// for human operators.
		// kafkaEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			// Keys can be anything except the empty string.
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.FullCallerEncoder,
		})

		cores = append(cores,
			zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
			zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
		)
	}

	// Join the outputs, encoders, and level-handling functions into
	// zapcore.Cores, then tee the four cores together.
	tree := zapcore.NewTee(cores...)

	log.opt.zopts = append(log.opt.zopts, zapFields(log.opt.tags))
	log.opt.zopts = append(log.opt.zopts, zap.AddCaller())
	log.opt.zopts = append(log.opt.zopts, zap.AddCallerSkip(1))

	log.logger = zap.New(tree, log.opt.zopts...)
	log.suger = log.logger.Sugar()

	// log.logger, _ = zap.NewProduction(zap.AddCallerSkip(1), zapFields(opt.tags))
	// log.suger = log.logger.Sugar()

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
		args = ex.WithArgs(ctx, args)
	}
	log.suger.Debug(args...)
}

// Debugfc .
func Debugfc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		format = ex.WithFormat(ctx, format)
	}
	log.suger.Debugf(format, args...)
}

// Debugwc .
func Debugwc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range log.opt.extractors {
		keysAndValues = ex.WithKVs(ctx, keysAndValues)
	}
	log.suger.Debugw(msg, keysAndValues...)
}

// Debugsc .
func Debugsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range log.opt.extractors {
		fields = ex.WithFields(ctx, fields)
	}
	log.logger.Debug(msg, fields...)
}

// Infoc .
func Infoc(ctx context.Context, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		args = ex.WithArgs(ctx, args)
	}
	log.suger.Info(args...)
}

// Infofc .
func Infofc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		format = ex.WithFormat(ctx, format)
	}
	log.suger.Infof(format, args...)
}

// Infowc .
func Infowc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range log.opt.extractors {
		keysAndValues = ex.WithKVs(ctx, keysAndValues)
	}
	log.suger.Infow(msg, keysAndValues...)
}

// Infosc .
func Infosc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range log.opt.extractors {
		fields = ex.WithFields(ctx, fields)
	}
	log.logger.Info(msg, fields...)
}

// Warnc .
func Warnc(ctx context.Context, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		args = ex.WithArgs(ctx, args)
	}
	log.suger.Warn(args...)
}

// Warnfc .
func Warnfc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		format = ex.WithFormat(ctx, format)
	}
	log.suger.Warnf(format, args...)
}

// Warnwc .
func Warnwc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range log.opt.extractors {
		keysAndValues = ex.WithKVs(ctx, keysAndValues)
	}
	log.suger.Warnw(msg, keysAndValues...)
}

// Warnsc .
func Warnsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range log.opt.extractors {
		fields = ex.WithFields(ctx, fields)
	}
	log.logger.Warn(msg, fields...)
}

// Errorc .
func Errorc(ctx context.Context, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		args = ex.WithArgs(ctx, args)
	}
	log.suger.Error(args...)
}

// Errorfc .
func Errorfc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		format = ex.WithFormat(ctx, format)
	}
	log.suger.Errorf(format, args...)
}

// Errorwc .
func Errorwc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range log.opt.extractors {
		keysAndValues = ex.WithKVs(ctx, keysAndValues)
	}
	log.suger.Errorw(msg, keysAndValues...)
}

// Errorsc .
func Errorsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range log.opt.extractors {
		fields = ex.WithFields(ctx, fields)
	}
	log.logger.Error(msg, fields...)
}

// Alarmsc .
func Alarmsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range log.opt.extractors {
		fields = ex.WithFields(ctx, fields)
	}
	var kvs []string
	for _, f := range fields {
		kvs = append(kvs, f.Key, sv(f))
	}
	if err := log.opt.alarmer.Alarm(ctx, "", msg, kvs...); err != nil {
		log.logger.Error("alarm err", zap.Error(err))
	}
	log.logger.Error(msg, fields...)
}

func sv(f zap.Field) string {
	switch f.Type {
	case zapcore.ArrayMarshalerType:
		return fmt.Sprintf("%v", f.Interface)
	case zapcore.ObjectMarshalerType:
		return fmt.Sprintf("%v", f.Interface)
	case zapcore.BinaryType:
		return fmt.Sprintf("%v", f.Interface)
	case zapcore.BoolType:
		return fmt.Sprintf("%v", f.Integer == 1)
	case zapcore.ByteStringType:
		return fmt.Sprintf("%v", f.Interface)
	case zapcore.Complex128Type:
		return fmt.Sprintf("%v", f.Interface)
	case zapcore.Complex64Type:
		return fmt.Sprintf("%v", f.Interface)
	case zapcore.DurationType:
		return fmt.Sprintf("%v", f.Integer)
	case zapcore.Float64Type:
		return fmt.Sprintf("%v", f.Integer)
	case zapcore.Float32Type:
		return fmt.Sprintf("%v", f.Integer)
	case zapcore.Int64Type:
		return fmt.Sprintf("%v", f.Integer)
	case zapcore.Int32Type:
		return fmt.Sprintf("%v", f.Integer)
	case zapcore.Int16Type:
		return fmt.Sprintf("%v", f.Integer)
	case zapcore.Int8Type:
		return fmt.Sprintf("%v", f.Integer)
	case zapcore.StringType:
		return f.String
	case zapcore.TimeType:
		return fmt.Sprintf("%v", f.Integer)
	case zapcore.Uint64Type:
		return fmt.Sprintf("%v", f.Integer)
	case zapcore.Uint32Type:
		return fmt.Sprintf("%v", f.Integer)
	case zapcore.Uint16Type:
		return fmt.Sprintf("%v", f.Integer)
	case zapcore.Uint8Type:
		return fmt.Sprintf("%v", f.Integer)
	case zapcore.UintptrType:
		return fmt.Sprintf("%v", f.Integer)
	case zapcore.ReflectType:
		return fmt.Sprintf("%v", f.Interface)
	case zapcore.NamespaceType:
		return f.Key
	case zapcore.StringerType:
		return fmt.Sprintf("%v", f.Interface)
	case zapcore.ErrorType:
		return fmt.Sprintf("%v", f.Interface)
	}
	return ""
}

// Panicc .
func Panicc(ctx context.Context, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		args = ex.WithArgs(ctx, args)
	}
	log.suger.Panic(args...)
}

// Panicfc .
func Panicfc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		format = ex.WithFormat(ctx, format)
	}
	log.suger.Panicf(format, args...)
}

// Panicwc .
func Panicwc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range log.opt.extractors {
		keysAndValues = ex.WithKVs(ctx, keysAndValues)
	}
	log.suger.Panicw(msg, keysAndValues...)
}

// Panicsc .
func Panicsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range log.opt.extractors {
		fields = ex.WithFields(ctx, fields)
	}
	log.logger.Panic(msg, fields...)
}

// Fatalc .
func Fatalc(ctx context.Context, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		args = ex.WithArgs(ctx, args)
	}
	log.suger.Fatal(args...)
}

// Fatalfc .
func Fatalfc(ctx context.Context, format string, args ...interface{}) {
	for _, ex := range log.opt.extractors {
		format = ex.WithFormat(ctx, format)
	}
	log.suger.Fatalf(format, args...)
}

// Fatalwc .
func Fatalwc(ctx context.Context, msg string, keysAndValues ...interface{}) {
	for _, ex := range log.opt.extractors {
		keysAndValues = ex.WithKVs(ctx, keysAndValues)
	}
	log.suger.Fatalw(msg, keysAndValues...)
}

// Fatalsc .
func Fatalsc(ctx context.Context, msg string, fields ...zap.Field) {
	for _, ex := range log.opt.extractors {
		fields = ex.WithFields(ctx, fields)
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
