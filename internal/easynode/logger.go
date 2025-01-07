package easynode

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _log *zap.SugaredLogger

func init() {
	_log = NewCustomLogger()
}

func NewCustomLogger() *zap.SugaredLogger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:     "time",
		LevelKey:    "level",
		MessageKey:  "msg",
		// CallerKey:   "caller",
		// EncodeLevel: func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) { enc.AppendString(level.CapitalString()) },
		// EncodeTime:  zapcore.ISO8601TimeEncoder,
		// EncodeCaller: zapcore.ShortCallerEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	)
	return zap.New(core, zap.AddCaller()).Sugar()
}
