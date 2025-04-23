package logger

import (
	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	logrus "github.com/sirupsen/logrus"
	"os"
)

type Logger struct {
	*logrus.Logger
}

func NewLogger() *Logger {
	l := &Logger{}
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableQuote:     true,
		PadLevelText:     true,
		QuoteEmptyFields: true,
		ForceColors:      true,
	})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stdout)

	hook := graylog.NewGraylogHook("gelf-udp-service:12222", map[string]interface{}{"service": "smart_image"})
	//w.CompressionType = graylog.NoCompress
	//w.CompressionLevel = 0
	//hook.SetWriter(w)
	//log.AddHook(hook)
	std := logrus.StandardLogger()
	std.AddHook(hook)
	l.Logger = std
	return l
}
