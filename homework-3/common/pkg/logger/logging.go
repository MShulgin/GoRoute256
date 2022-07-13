package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func init() {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.StacktraceKey = ""
	config := zap.NewProductionConfig()
	config.EncoderConfig = encoderConfig
	var err error
	Log, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
}

func Info(msg string, vars ...zap.Field) {
	Log.Info(msg, vars...)
}

func Error(msg string, vars ...zap.Field) {
	Log.Error(msg, vars...)
}
