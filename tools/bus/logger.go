package bus

import "go.uber.org/zap"

type queueLogger struct {
	l *zap.Logger
}

func (q *queueLogger) Log(args ...interface{}) {
	q.l.Sugar().Debug(args...)
}
