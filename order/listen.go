package order

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coinbase-samples/ib-venue-listener-go/cloud"
	ws "github.com/coinbase-samples/ib-venue-listener-go/websocket"

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
	var ud = &Update{}
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
		c, err := ws.DialWebSocket(context.TODO(), app)
		if err != nil {
			log.Error(err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err := c.WriteMessage(websocket.TextMessage, []byte(subscribeOrdersString(app))); err != nil {
			log.Errorf("Unable to subscribe: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err := processMessages(app, c); err != nil {
			log.Error(err)
		}
	}
}

func subscribeOrdersString(app config.AppConfig) string {
	msgType := "subscribe"
	channel := "orders"
	key := app.AccessKey
	accountId := app.SenderId
	portfolioId := app.PortfolioId
	dt := time.Now().UTC()
	msgTime := dt.Format(time.RFC3339)
	signature := prime.Sign(channel, key, accountId, msgTime, portfolioId, "", app.SigningKey)
	message := fmt.Sprintf(`{
        "type": "%s",
        "channel": "%s",
        "access_key": "%s",
        "api_key_id": "%s",
				"portfolio_id": "%s",
        "signature": "%s",
        "passphrase": "%s",
        "timestamp": "%s"
      }`, msgType, channel, key, accountId, portfolioId, signature, app.Passphrase, msgTime)
	return message
}

func writeOrderUpdatesToEventBus(
	app config.AppConfig,
	orderUpdate *Update,
) {
	// loop in loop because everything is an array for some reason
	for _, event := range orderUpdate.Events {
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
