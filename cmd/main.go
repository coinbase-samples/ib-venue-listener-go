package main

import (
	"github.com/coinbase-samples/ib-venue-listener-go/config"
	"github.com/coinbase-samples/ib-venue-listener-go/order"
	"github.com/coinbase-samples/ib-venue-listener-go/prices"
	log "github.com/sirupsen/logrus"
)

var (
	logrusLogger = log.New()
)

func main() {

	var app config.AppConfig

	if err := config.Setup(&app); err != nil {
		log.Fatalf("Unable to config app: %v", err)
	}

	// This will print the prime credentials
	if app.IsLocalEnv() {
		log.Debugf("starting app with config: %v", app)
	}

	logLevel, _ := log.ParseLevel(app.LogLevel)
	log.SetLevel(logLevel)

	order.StartListener(app)
	prices.StartListener(app)
}
