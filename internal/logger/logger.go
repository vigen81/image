package logger

import (
	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	Log *log.Logger
)

func init() {

	log.SetFormatter(&log.TextFormatter{
		DisableQuote:     true,
		PadLevelText:     true,
		QuoteEmptyFields: true,
		ForceColors:      true,
	})
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)

	hook := graylog.NewGraylogHook("gelf-udp-service:12222", map[string]interface{}{"service": "smart_image"})
	//w.CompressionType = graylog.NoCompress
	//w.CompressionLevel = 0
	//hook.SetWriter(w)
	//log.AddHook(hook)
	Log = log.StandardLogger()
	Log.AddHook(hook)

}
