package order

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/coinbase-samples/ib-venue-listener-go/cloud"
	"github.com/recws-org/recws"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/coinbase-samples/ib-venue-listener-go/config"
	"github.com/coinbase-samples/ib-venue-listener-go/prime"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

func StartListener(app config.AppConfig) {
	//Create Message Out
	messageOut := make(chan string)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: app.PrimeApiUrl}

	log.Printf("connecting to %s", u.String())
	ctx, _ := context.WithCancel(context.Background())
	ws := recws.RecConn{}
	ws.Dial(u.String(), nil)

	//When the program closes, close the connection
	defer ws.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		messageOut <- subscribeOrdersString(app)
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Error("read:", err)
				return
			}

			var ud = &Update{}
			err = json.Unmarshal(message, &ud)
			writeOrderUpdatesToEventBus(app, ud)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Websocket closed %s", ws.GetURL())
			return
		case m := <-messageOut:
			log.Printf("Send Message %s", m)
			err := ws.WriteMessage(websocket.TextMessage, []byte(m))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-ctx.Done():
			case <-time.After(time.Second):
			}
			return
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
