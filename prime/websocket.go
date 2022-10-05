package prime

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/coinbase-samples/ib-venue-listener-go/config"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

func WriteCloseMessge(c *websocket.Conn) {
	if err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
		log.Errorf("Write close: %v", err)
	}
}

func HeartbeatSubscriptionMsg(app config.AppConfig) []byte {
	msgType := "subscribe"
	channel := "heartbeat"
	key := app.AccessKey
	accountId := app.SenderId
	portfolioId := app.PortfolioId
	dt := time.Now().UTC()
	msgTime := dt.Format(time.RFC3339)

	signature := Sign(channel, key, accountId, msgTime, portfolioId, "", app.SigningKey)

	return []byte(fmt.Sprintf(`{
        "type": "%s",
        "channel": "%s",
        "access_key": "%s",
        "api_key_id": "%s",
				"portfolio_id": "%s",
        "signature": "%s",
        "passphrase": "%s",
        "timestamp": "%s",
				"prouduct_ids": "[\"BTC-USD\"]",
      }`, msgType, channel, key, accountId, portfolioId, signature, app.Passphrase, msgTime))
}

func DialWebSocket(ctx context.Context, app config.AppConfig) (*websocket.Conn, error) {

	u := url.URL{Scheme: "wss", Host: app.PrimeApiUrl}

	log.Debugf("Connecting to %s", u.String())

	// TODO: Do we need to look at the response status/code?
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Cannot dial WebSocket: %v", err)
	}

	return c, nil
}
