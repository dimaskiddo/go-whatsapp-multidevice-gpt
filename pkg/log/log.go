package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

type logLevel string

const (
	LogLevelPanic logLevel = "panic"
	LogLevelFatal logLevel = "fatal"
	LogLevelError logLevel = "error"
	LogLevelWarn  logLevel = "warn"
	LogLevelDebug logLevel = "debug"
	LogLevelTrace logLevel = "trace"
	LogLevelInfo  logLevel = "info"
)

func init() {
	logger = logrus.New()

	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
}

func Println(level logLevel, message interface{}) {
	if logger != nil {
		switch level {
		case "panic":
			logger.Panicln(message)
		case "fatal":
			logger.Fatalln(message)
		case "error":
			logger.Errorln(message)
		case "warn":
			logger.Warnln(message)
		case "debug":
			logger.Debugln(message)
		case "trace":
			logger.Traceln(message)
		default:
			logger.Infoln(message)
		}
	}
}
