package config

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func LogInit(app AppConfig, f *os.File) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)
	logLevel, _ := log.ParseLevel(app.LogLevel)
	log.SetLevel(logLevel)

	if app.LogToFile == "true" {
		log.SetOutput(f)
	} else {
		log.SetOutput(os.Stdout)
	}
}
