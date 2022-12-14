package config

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func LogInit(app AppConfig) *os.File {
	logger := log.New()
	logLevel, _ := log.ParseLevel(app.LogLevel)
	logger.SetLevel(logLevel)
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetReportCaller(true)
	var f *os.File
	if app.LogToFile == "true" {
		// open a file
		f, err := os.OpenFile("testing.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("error opening file: %v", err)
		}
		log.SetOutput(f)
	} else {
		log.SetOutput(os.Stdout)
	}

	return f
}
