package prices

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/coinbase-samples/ib-venue-listener-go/cloud"
	"github.com/coinbase-samples/ib-venue-listener-go/config"
	"github.com/coinbase-samples/ib-venue-listener-go/prime"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

func StartListener(app config.AppConfig) {
	//Create Message Out
	messageOut := make(chan string)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: app.PrimeApiUrl}

	log.Infof("Connecting to %s", u.String())

	c, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		// NOTE: resp is nil if there is an error
		//log.Printf("handshake failed with status %d", resp.StatusCode)
		log.Fatalf("Cannot dial WebSocket: %v", err)
	}

	// TODO: What is the expected response status code?
	if resp == nil {

	}

	summary := PriceSummary{
		Assets: []AssetPrice{
			{
				Name:      "Bitcoin",
				Ticker:    "BTC-USD",
				HighOffer: math.NaN(),
				LowBid:    math.NaN(),
				Spread:    math.NaN(),
			}, {
				Name:      "Ethereum",
				Ticker:    "ETH-USD",
				HighOffer: math.NaN(),
				LowBid:    math.NaN(),
				Spread:    math.NaN(),
			}, {
				Name:      "Solana",
				Ticker:    "SOL-USD",
				HighOffer: math.NaN(),
				LowBid:    math.NaN(),
				Spread:    math.NaN(),
			}, {
				Name:      "Cosmos",
				Ticker:    "ATOM-USD",
				HighOffer: math.NaN(),
				LowBid:    math.NaN(),
				Spread:    math.NaN(),
			}, {
				Name:      "Polygon",
				Ticker:    "MATIC-USD",
				HighOffer: math.NaN(),
				LowBid:    math.NaN(),
				Spread:    math.NaN(),
			}, {
				Name:      "Cardano",
				Ticker:    "ADA-USD",
				HighOffer: math.NaN(),
				LowBid:    math.NaN(),
				Spread:    math.NaN(),
			},
		}}

	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, asset := range summary.Assets {

					writeAssetPriceToEventBus(app, asset)

				}

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	// When the program closes close the connection
	defer c.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		messageOut <- subscribePricesString(app)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Errorf("Read message:", err)
				return
			}

			//log.Infof("recv: %s", message)
			var ud = &OrderBookUpdate{}
			json.Unmarshal(message, &ud)

			if len(ud.Events) > 0 && len(ud.Events[0].Updates) > 0 {
				for _, row := range ud.Events {
					product := row.ProductID
					assetPriceIdx := slices.IndexFunc(summary.Assets, func(a AssetPrice) bool { return a.Ticker == product })
					assetPrice := summary.Assets[assetPriceIdx]
					//log.Printf("existing summary - %v", assetPrice)

					floor, ceiling, spread := float64(1000000), float64(0), float64(0)

					for _, row := range row.Updates {

						if row.Qty == "0" {
							//skip
						} else {
							rowPrice, _ := strconv.ParseFloat(row.Px, 32)
							if row.Side == "offer" && rowPrice < floor {
								floor = rowPrice
							} else if row.Side == "bid" && rowPrice > ceiling {

								ceiling = rowPrice
							}
						}
					}
					spread = ceiling - floor

					summary.Assets[assetPriceIdx] = AssetPrice{
						Name:      assetPrice.Name,
						Ticker:    assetPrice.Ticker,
						HighOffer: ceiling,
						LowBid:    floor,
						Spread:    spread,
					}
				}

			}
		}

	}()

	for {
		select {
		case <-done:
			return
		case m := <-messageOut:
			log.Printf("Send Message %s", m)
			err := c.WriteMessage(websocket.TextMessage, []byte(m))
			if err != nil {
				log.Errorf("Write: %v", err)
				return
			}
		case <-interrupt:
			log.Debug("Interrupt received")
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Infof("Write close: %v", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func writeAssetPriceToEventBus(
	app config.AppConfig,
	asset AssetPrice,
) {

	if asset.NotSet() {
		return
	}

	val, err := json.Marshal(asset)
	if err != nil {
		log.Errorf("Unable to marshal asset: %v", err)
		return
	}

	val = append(val, []byte("\n")...)

	dst := make([]byte, base64.StdEncoding.EncodedLen(len(val)))
	base64.StdEncoding.Encode(dst, val)

	if err := cloud.KdsPutRecord(context.Background(), app, dst); err != nil {

	}

}

func subscribePricesString(app config.AppConfig) string {
	msgType := "subscribe"
	channel := "l2_data"
	key := app.AccessKey
	accountId := app.SenderId

	productIds := `["BTC-USD", "ETH-USD", "ADA-USD", "MATIC-USD", "ATOM-USD", "SOL-USD"]`
	//productIds := `["BTC-USD"]`

	t := time.Now()
	msgTime := t.UTC().Format(time.RFC3339)

	signature := prime.Sign(channel, key, accountId, msgTime, "", productIds, app.SigningKey)

	message := fmt.Sprintf(`{
		"type": "%s",
		"channel": "%s",
		"product_ids": %s,
		"access_key": "%s",
		"api_key_id": "%s",
		"signature": "%s",
		"passphrase": "%s",
		"timestamp": "%s" }`,
		msgType, channel, productIds, key, accountId, signature, app.Passphrase, msgTime)

	if app.IsLocalEnv() {
		//log.Debug("message %v", message)

	}

	return message
}
