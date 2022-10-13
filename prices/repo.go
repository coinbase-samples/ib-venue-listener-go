package prices

import (
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Repo the repository used by dynamo
var Repo *Repository

// Repository is the repository type
type Repository struct {
	AssetCache Cache
}

// NewRepo creates a new repository
func NewRepo(c Cache) *Repository {

	return &Repository{
		AssetCache: c,
	}
}

// NewDBA sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (m *Repository) HandlePriceUpdate(ctx context.Context, assetPrice AssetPrice) {
	assets, err := m.AssetCache.GetAssets(ctx)
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

	if err := PutAsset(ctx, asset); err != nil {
		log.Error(err)
	}

}

func findAsset(primeProductId string, assets []Asset) *Asset {
	// TODO: Move Asset table to use the Prime product id / ticker.
	ticker := strings.Replace(primeProductId, "-USD", "", 1)

	for _, a := range assets {
		if a.Ticker == ticker {
			return &a
		}
	}

	return nil
}
