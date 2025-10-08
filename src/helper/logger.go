package helper

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func InitLogger() {
	Log = logrus.New()
	Log.SetFormatter(&logrus.JSONFormatter{})
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		Log.SetLevel(logrus.InfoLevel)
	} else {
		parsedLevel, err := logrus.ParseLevel(level)
		if err != nil {
			Log.SetLevel(logrus.InfoLevel)
		} else {
			Log.SetLevel(parsedLevel)
		}
	}
}
