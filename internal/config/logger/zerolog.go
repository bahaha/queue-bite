package log

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type ZerologLogger struct {
	log zerolog.Logger
}

func NewZerologLogger(w io.Writer, isDev bool) Logger {
	output := zerolog.ConsoleWriter{
		Out:        w,
		TimeFormat: time.RFC3339,
		NoColor:    !isDev,
		FormatLevel: func(i interface{}) string {
			return "[" + strings.ToUpper(i.(string)) + "]"
		},
	}

	log := zerolog.New(output).With().Timestamp().Logger()

	return &ZerologLogger{
		log: log,
	}
}

func (zl *ZerologLogger) LogDebug(component, msg string, args ...interface{}) {
	event := zl.log.Debug()
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			event.Interface(args[i].(string), args[i+1])
		}
	}
	event.Msg(fmt.Sprintf("(%s) %s", component, msg))
}

func (zl *ZerologLogger) LogInfo(component, msg string, args ...interface{}) {
	event := zl.log.Info()
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			event.Interface(args[i].(string), args[i+1])
		}
	}
	event.Msg(fmt.Sprintf("(%s) %s", component, msg))
}

func (zl *ZerologLogger) LogErr(component string, err error, msg string, args ...interface{}) {
	event := zl.log.Error().Err(err)
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			event.Interface(args[i].(string), args[i+1])
		}
	}
	event.Msg(fmt.Sprintf("(%s) %s", component, msg))
}
