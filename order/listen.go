package order

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coinbase-samples/ib-venue-listener-go/cloud"

	"github.com/coinbase-samples/ib-venue-listener-go/config"
	"github.com/coinbase-samples/ib-venue-listener-go/prime"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

func RunListener(app config.AppConfig) {

	// TODO: Implement a context for cancel / shutdown

	processMessagesWithReconnect(app)
}

func processMessage(app config.AppConfig, message []byte) error {
	var ud = &prime.OrderUpdate{}
	if err := json.Unmarshal(message, ud); err != nil {
		return fmt.Errorf("Unable to umarshal json: %s - msg: %v", string(message), err)
	}

	writeOrderUpdatesToEventBus(app, ud)

	return nil
}

func processMessages(app config.AppConfig, c *websocket.Conn) error {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return fmt.Errorf("Problem reading msg: %v", err)
		}

		if err := processMessage(app, message); err != nil {
			return err
		}
	}
}

func processMessagesWithReconnect(app config.AppConfig) {
	for {
		c, err := prime.DialWebSocket(context.TODO(), app)
		if err != nil {
			log.Error(err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err := c.WriteMessage(websocket.TextMessage, subscribeOrdersString(app)); err != nil {
			log.Errorf("Unable to subscribe to orders feed: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err := c.WriteMessage(websocket.TextMessage, prime.HeartbeatSubscriptionMsg(app)); err != nil {
			log.Errorf("Unable to subscribe to heartbeats: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err := processMessages(app, c); err != nil {
			log.Error(err)
		}
	}
}

func subscribeOrdersMsg(app config.AppConfig) []byte {
	msgType := "subscribe"
	channel := "orders"
	key := app.AccessKey
	accountId := app.SenderId
	portfolioId := app.PortfolioId
	dt := time.Now().UTC()
	msgTime := dt.Format(time.RFC3339)
	signature := prime.Sign(channel, key, accountId, msgTime, portfolioId, "", app.SigningKey)
	return []byte(fmt.Sprintf(`{
        "type": "%s",
        "channel": "%s",
        "access_key": "%s",
        "api_key_id": "%s",
				"portfolio_id": "%s",
        "signature": "%s",
        "passphrase": "%s",
        "timestamp": "%s"
      }`, msgType, channel, key, accountId, portfolioId, signature, app.Passphrase, msgTime))
}

func writeOrderUpdatesToEventBus(
	app config.AppConfig,
	orderUpdate *prime.OrderUpdate,
) {
	// loop in loop because everything is an array for some reason
	for _, event := range orderUpdate.Events {

		if event.Type == "error" {
			if len(event.Message) > 0 {
				log.Errorf("Orders channel error: %s", event.Message)
			}
			continue
		}

		for _, order := range event.Orders {
			val, err := json.Marshal(order)
			if err != nil {
				log.Errorf("Unable to marshal asset: %v", err)
				return
			}

			val = append(val, []byte("\n")...)

			dst := make([]byte, base64.StdEncoding.EncodedLen(len(val)))
			base64.StdEncoding.Encode(dst, val)

			if err := cloud.KdsPutRecord(
				context.Background(),
				app,
				app.OrderKinesisStreamName,
				order.ClientOrderID,
				dst,
			); err != nil {
				log.Errorf("Unable to put KDS record: %v", err)
			}
		}
	}
}
