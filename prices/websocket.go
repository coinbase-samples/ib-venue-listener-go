package prices

import (
	"context"
	"fmt"
	"net/url"

	"github.com/coinbase-samples/ib-venue-listener-go/config"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

func writeCloseMessge(c *websocket.Conn) {
	if err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
		log.Errorf("Write close: %v", err)
	}
}

func dialWebSocket(ctx context.Context, app config.AppConfig) (*websocket.Conn, error) {

	u := url.URL{Scheme: "wss", Host: app.PrimeApiUrl}

	log.Debugf("Connecting to %s", u.String())

	// TODO: Do we need to look at the response status/code?
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Cannot dial WebSocket: %v", err)
	}

	return c, nil
}
