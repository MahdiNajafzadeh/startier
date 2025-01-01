package easynode

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _log *zap.SugaredLogger

func init() {
	_log = NewCustomLogger()
}

func NewCustomLogger() *zap.SugaredLogger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:    "time",
		LevelKey:   "level",
		MessageKey: "msg",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006/01/02 15:04:05"))
		},
		EncodeLevel:      func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) { enc.AppendString(level.CapitalString()) },
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		// ConsoleSeparator: " . ",
	}
	return zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			os.Stdout,
			zapcore.DebugLevel,
		)).Sugar()
	// l, err := zap.NewDevelopment()
	// if err != nil {
	// 	panic(err)
	// }
	// return l.Sugar()
}
