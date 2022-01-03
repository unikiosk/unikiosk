package logger

import "go.uber.org/zap"

type ProxyLogger interface {
	Printf(format string, v ...interface{})
}

type proxyLogger struct {
	log *zap.Logger
}

func NewProxyLogger(log *zap.Logger) *proxyLogger {
	return &proxyLogger{
		log: log,
	}
}

func (l *proxyLogger) Printf(format string, v ...interface{}) {
	l.log.Sugar().Infof(format, v...)
}
