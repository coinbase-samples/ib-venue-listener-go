package main

import (
	"fmt"

	"github.com/coinbase-samples/ib-venue-listener-go/config"
	"github.com/coinbase-samples/ib-venue-listener-go/prices"
	log "github.com/sirupsen/logrus"
)

var (
	logrusLogger = log.New()
)

func main() {

	var app config.AppConfig

	config.Setup(&app)
	fmt.Println("starting app with config", app)

	logLevel, _ := log.ParseLevel(app.LogLevel)
	logrusLogger.SetLevel(logLevel)
	logrusLogger.SetFormatter(&log.JSONFormatter{})

	//order.StartListener(app)
	prices.StartListener(app)
}
