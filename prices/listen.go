package prices

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coinbase-samples/ib-venue-listener-go/config"
	"github.com/coinbase-samples/ib-venue-listener-go/dba"
	"github.com/coinbase-samples/ib-venue-listener-go/model"
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

func ProcessOrderBookUpdates(message []byte) error {
	var ud = &model.OrderBookUpdate{}
	if err := json.Unmarshal(message, ud); err != nil {
		return fmt.Errorf("unable to umarshal json: %s - msg: %w", string(message), err)
	}

	for _, row := range ud.Events {

		if row.Type == "error" {
			if len(row.Message) > 0 {
				log.Errorf("Prices channel error: %s", row.Message)
			}
			continue
		}

		assetPriceIdx := slices.IndexFunc(
			prices.Assets,
			func(a model.AssetPrice) bool { return a.Ticker == row.ProductID },
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

			rowPrice, _ := strconv.ParseFloat(row.Px, 64)

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

		prices.Assets[assetPriceIdx] = model.AssetPrice{
			Name:      assetPrice.Name,
			Ticker:    assetPrice.Ticker,
			HighOffer: ceiling,
			LowBid:    floor,
			Spread:    ceiling - floor,
		}
	}

	return nil
}

func EmitPriceUpdates(app config.AppConfig, interrupt chan os.Signal) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			for _, asset := range prices.Assets {
				HandlePriceUpdate(context.Background(), asset)
			}
		case <-interrupt:
			log.Debugln("stopping price emitter")
			ticker.Stop()
		}

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
