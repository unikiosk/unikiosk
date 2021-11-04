package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ParseLogLevel(level string) zapcore.Level {
	switch level {
	case "info":
		return zap.InfoLevel
	case "debug":
		return zap.DebugLevel
	default:
		return zap.InfoLevel
	}
}
