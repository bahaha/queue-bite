package log

const (
	Server = "sys/server"
)

type Logger interface {
	LogDebug(component, msg string, args ...interface{})
	LogInfo(component, msg string, args ...interface{})
	LogErr(component string, err error, msg string, args ...interface{})
}