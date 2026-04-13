package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	l *zap.Logger
}

func NewZapLogger() *ZapLogger {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Encoding:    "console",
		Development: true,

		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:      "time",
			LevelKey:     "level",
			MessageKey:   "msg",
			CallerKey:    "caller",
			EncodeLevel:  zapcore.CapitalColorLevelEncoder,
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},

		OutputPaths: []string{"stdout"},
	}

	z, err := cfg.Build(zap.AddCaller(),
		zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
	return &ZapLogger{l: z}
}

func (z *ZapLogger) Debug(msg string, fields ...Field) {
	z.l.Debug(msg, toZapFields(fields)...)
}

func (z *ZapLogger) Info(msg string, fields ...Field) {
	z.l.Info(msg, toZapFields(fields)...)
}

func (z *ZapLogger) Warn(msg string, fields ...Field) {
	z.l.Warn(msg, toZapFields(fields)...)
}

func (z *ZapLogger) Error(msg string, fields ...Field) {
	z.l.Error(msg, toZapFields(fields)...)
}

func (z *ZapLogger) Panic(msg string, fields ...Field) {
	z.l.Panic(msg, toZapFields(fields)...)
}

func (z *ZapLogger) Close() {
	if err := z.l.Sync(); err != nil {
		panic(err)
	}
}

func toZapFields(fields []Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	res := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		res = append(res, zap.Any(f.Key, f.Value))
	}

	return res
}
