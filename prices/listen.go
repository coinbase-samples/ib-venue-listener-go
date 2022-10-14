package prices

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/coinbase-samples/ib-venue-listener-go/cloud"
	"github.com/coinbase-samples/ib-venue-listener-go/config"
	"github.com/coinbase-samples/ib-venue-listener-go/dba"
	"github.com/coinbase-samples/ib-venue-listener-go/model"
	"github.com/coinbase-samples/ib-venue-listener-go/prime"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

var prices = model.PriceSummary{
	Assets: []model.AssetPrice{
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
	productIds := `["BTC-USD", "ETH-USD", "ADA-USD", "MATIC-USD", "ATOM-USD", "SOL-USD"]`

	for {
		c, err := prime.DialWebSocket(context.TODO(), app)
		if err != nil {
			log.Error(err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Debugf("starting price subscription to -%v", productIds)
		if err := c.WriteMessage(websocket.TextMessage, prime.PricesSubscriptionMsg(app, productIds)); err != nil {
			log.Errorf("Unable to subscribe to price feed: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Debugf("starting heartbeat subscription")
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

func processOrderBookUpdates(ud *model.OrderBookUpdate) {
	for _, row := range ud.Events {

		if row.Type == "error" {
			if len(row.Message) > 0 {
				log.Errorf("Prices channel error: %s", row.Message)
			}
			continue
		}

		product := row.ProductID

		assetPriceIdx := slices.IndexFunc(
			prices.Assets,
			func(a model.AssetPrice) bool { return a.Ticker == product },
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

		prices.Assets[assetPriceIdx] = model.AssetPrice{
			Name:      assetPrice.Name,
			Ticker:    assetPrice.Ticker,
			HighOffer: ceiling,
			LowBid:    floor,
			Spread:    spread,
		}
	}
}

func processMessage(message []byte) error {
	var ud = &model.OrderBookUpdate{}
	if err := json.Unmarshal(message, ud); err != nil {
		return fmt.Errorf("unable to umarshal json: %s - msg: %v", string(message), err)
	}

	processOrderBookUpdates(ud)

	return nil
}

func processMessages(app config.AppConfig, c *websocket.Conn) error {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return fmt.Errorf("problem reading msg: %v", err)
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
				//writeAssetPriceToEventBus(app, asset)
				HandlePriceUpdate(context.TODO(), asset)
			}
		}
	}
}

func writeAssetPriceToEventBus(
	app config.AppConfig,
	asset model.AssetPrice,
) {
	if asset.NotSet() {
		return
	}

	if !app.IsLocalEnv() {
		HandlePriceUpdate(context.TODO(), asset)
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

func HandlePriceUpdate(ctx context.Context, assetPrice model.AssetPrice) {
	assets, err := dba.Repo.Cache.GetAssets(ctx)
	if err != nil {
		log.Errorf("Unable to fetch assets: %v", err)
		return
	}

	asset := findAsset(assetPrice.Ticker, assets)
	if asset == nil {
		log.Debugf("Unable to find prouct: %s", assetPrice.Ticker)
		return
	}

	// TODO: We should be storing in atomic units
	asset.HighOffer = fmt.Sprintf("%f", assetPrice.HighOffer)
	asset.LowBid = fmt.Sprintf("%f", assetPrice.LowBid)
	asset.Spread = fmt.Sprintf("%f", assetPrice.HighOffer-assetPrice.LowBid)

	log.Debugf("updating asset - %v", asset)
	if err := dba.Repo.PutAsset(ctx, asset); err != nil {
		log.Error(err)
	}

}

func findAsset(primeProductId string, assets []model.Asset) *model.Asset {
	// TODO: Move Asset table to use the Prime product id / ticker.
	ticker := strings.Replace(primeProductId, "-USD", "", 1)

	for _, a := range assets {
		if a.Ticker == ticker {
			return &a
		}
	}

	return nil
}
