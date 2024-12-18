package log

type NoopLogger struct {
}

func NewNoopLogger() Logger {
	return &NoopLogger{}
}

func (l *NoopLogger) LogDebug(component, msg string, args ...interface{}) {
}

func (l *NoopLogger) LogInfo(component, msg string, args ...interface{}) {
}

func (l *NoopLogger) LogErr(component string, err error, msg string, args ...interface{}) {
}
