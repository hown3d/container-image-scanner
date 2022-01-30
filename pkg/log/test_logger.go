package log

import "testing"

type TestLogger struct {
	T *testing.T
}

func (l TestLogger) Error(args ...interface{}) {
	l.T.Error(args)
}
func (l TestLogger) Errorf(format string, args ...interface{}) {
	l.T.Errorf(format, args)
}
func (l TestLogger) Info(args ...interface{}) {
	l.T.Log(args)
}
func (l TestLogger) Infof(format string, args ...interface{}) {
	l.T.Logf(format, args)
}
func (l TestLogger) Debug(args ...interface{}) {
	l.Info(args)
}
func (l TestLogger) Debugf(format string, args ...interface{}) {
	l.Infof(format, args)
}
