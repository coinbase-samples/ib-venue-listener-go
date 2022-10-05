package prices

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/coinbase-samples/ib-venue-listener-go/cloud"
	"github.com/coinbase-samples/ib-venue-listener-go/config"
	"github.com/coinbase-samples/ib-venue-listener-go/prime"
	ws "github.com/coinbase-samples/ib-venue-listener-go/websocket"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

var prices = PriceSummary{
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

func RunListener(app config.AppConfig) {

	// TODO: Implement a context for cancel / shutdown

	go emitPriceUpdates(app)

	processMessagesWithReconnect(app)

}

func processMessagesWithReconnect(app config.AppConfig) {
	for {
		c, err := ws.DialWebSocket(context.TODO(), app)
		if err != nil {
			log.Error(err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err := c.WriteMessage(websocket.TextMessage, []byte(subscribePricesString(app))); err != nil {
			log.Errorf("Unable to subscribe: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err := processMessages(app, c); err != nil {
			log.Error(err)
		}
	}
}

func processOrderBookUpdate(ud *OrderBookUpdate) {
	for _, row := range ud.Events {

		product := row.ProductID

		assetPriceIdx := slices.IndexFunc(
			prices.Assets,
			func(a AssetPrice) bool { return a.Ticker == product },
		)

		if assetPriceIdx == -1 {
			continue
		}

		assetPrice := prices.Assets[assetPriceIdx]

		floor, ceiling := math.NaN(), math.NaN()

		for _, row := range row.Updates {

			if row.Qty == "0" {
				continue
			}

			rowPrice, _ := strconv.ParseFloat(row.Px, 32)

			if row.Side == "offer" && (math.IsNaN(floor) || rowPrice < floor) {

				floor = rowPrice

			} else if row.Side == "bid" && (math.IsNaN(ceiling) || rowPrice > ceiling) {

				ceiling = rowPrice
			}
		}

		if math.IsNaN(ceiling) && !math.IsNaN(prices.Assets[assetPriceIdx].HighOffer) {
			ceiling = prices.Assets[assetPriceIdx].HighOffer
		}

		if math.IsNaN(floor) && !math.IsNaN(prices.Assets[assetPriceIdx].LowBid) {
			floor = prices.Assets[assetPriceIdx].LowBid
		}

		spread := ceiling - floor

		prices.Assets[assetPriceIdx] = AssetPrice{
			Name:      assetPrice.Name,
			Ticker:    assetPrice.Ticker,
			HighOffer: ceiling,
			LowBid:    floor,
			Spread:    spread,
		}
	}
}

func processMessage(message []byte) error {
	var ud = &OrderBookUpdate{}
	if err := json.Unmarshal(message, ud); err != nil {
		return fmt.Errorf("Unable to umarshal json: %s - msg: %v", string(message), err)
	}

	processOrderBookUpdate(ud)

	return nil
}

func processMessages(app config.AppConfig, c *websocket.Conn) error {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return fmt.Errorf("Problem reading msg: %v", err)
		}

		if err := processMessage(message); err != nil {
			return err
		}
	}
}

func emitPriceUpdates(app config.AppConfig) {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			for _, asset := range prices.Assets {
				writeAssetPriceToEventBus(app, asset)
			}
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

	if err := cloud.KdsPutRecord(
		context.Background(),
		app,
		app.PriceKinesisStreamName,
		asset.Ticker,
		dst,
	); err != nil {
		log.Errorf("Unable to put KDS record: %v", err)
	}
}

func subscribePricesString(app config.AppConfig) string {
	msgType := "subscribe"
	channel := "l2_data"
	key := app.AccessKey
	accountId := app.SenderId

	productIds := `["BTC-USD", "ETH-USD", "ADA-USD", "MATIC-USD", "ATOM-USD", "SOL-USD"]`

	msgTime := time.Now().UTC().Format(time.RFC3339)

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

	return message
}
