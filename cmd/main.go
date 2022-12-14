package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/coinbase-samples/ib-venue-listener-go/config"
	"github.com/coinbase-samples/ib-venue-listener-go/dba"
	"github.com/coinbase-samples/ib-venue-listener-go/model"
	"github.com/coinbase-samples/ib-venue-listener-go/order"
	"github.com/coinbase-samples/ib-venue-listener-go/prices"
	"github.com/coinbase-samples/ib-venue-listener-go/prime"
	"github.com/coinbase-samples/ib-venue-listener-go/queue"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
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

	f := config.LogInit(app)

	if f != nil {
		defer f.Close()
	}

	dba.NewDBA(dba.NewRepo(&app))
	queue.NewQueue(queue.NewRepo(&app))

	run := make(chan os.Signal, 1)
	signal.Notify(run, os.Interrupt)

	// start websocket listener
	go processMessagesWithReconnect(app, run)
	// start price emitter
	go prices.EmitPriceUpdates(app, run)

	<-run
}

func processMessagesWithReconnect(app config.AppConfig, interrupt chan os.Signal) {

	done := make(chan struct{})
	for {
		log.Infof("connecting websocket to %s", app.PrimeApiUrl)
		c, err := prime.DialWebSocket(context.Background(), app)

		if err != nil {
			log.Error(err)
			time.Sleep(2 * time.Second)
			continue
		}
		defer func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("Recovered in closing ws", r)
				}
			}()
			c.Close()
		}()

		sendSubscribeMessages(app, c)
		if err := processMessages(app, c, done); err != nil {
			log.Error(err)
			continue
		}

		for {
			select {
			case <-done:
				return
			case <-interrupt:
				log.Println("interrupt")

				select {
				case <-done:
				case <-time.After(time.Second):
				}
				return
			}
		}

	}

}

func sendSubscribeMessages(app config.AppConfig, c *websocket.Conn) error {
	log.Debugf("starting heartbeat subscription")
	if err := c.WriteMessage(websocket.TextMessage, prime.HeartbeatSubscriptionMsg(app)); err != nil {
		return fmt.Errorf("unable to subscribe to heartbeats: %w", err)
	}
	log.Debugf("sent heartbeat subscription")

	log.Debugf("starting price subscription to - %v", app.ProductIds)
	if err := c.WriteMessage(websocket.TextMessage, prime.PricesSubscriptionMsg(app)); err != nil {
		return fmt.Errorf("unable to subscribe to price feed: %w", err)
	}
	log.Debugf("sent price subscription")

	log.Debugln("starting order subscription")
	if err := c.WriteMessage(websocket.TextMessage, prime.OrderSubscriptionMsg(app)); err != nil {
		return fmt.Errorf("unable to subscribe to orders feed: %w", err)
	}
	log.Debugln("started order subscription")
	return nil
}

func processMessage(app config.AppConfig, message []byte) error {
	var ud = &model.GenericMessage{}
	if err := json.Unmarshal(message, ud); err != nil {
		return fmt.Errorf("unable to umarshal json: %s - msg: %v", string(message), err)
	}

	// process by channel
	if ud.Channel == "l2_data" {
		prices.ProcessOrderBookUpdates(message)
	} else if ud.Channel == "subscriptions" {
		log.Debugf("subscription message - %s", string(message))
		var hd = &model.HeartbeatMessage{}
		if err := json.Unmarshal(message, hd); err != nil {
			return fmt.Errorf("unable to umarshal json: %s - msg: %v", string(message), err)
		}
		log.Debugf("parsed subscription - %v", hd)
	} else if ud.Channel == "heartbeat" {
		log.Debugf("heartbeat incoming! - %s", string(message))
	} else if ud.Channel == "orders" {
		order.ProcessOrderMessage(app, message)
	}

	return nil
}

func processMessages(app config.AppConfig, c *websocket.Conn, done chan struct{}) error {
	defer close(done)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Errorf("error reading message: %v - websocket - %v", message, c)
			continue
		}

		if err := processMessage(app, message); err != nil {
			log.Errorf("error processing message: %v - websocket - %v", message, c)
			return err
		}
	}
}
