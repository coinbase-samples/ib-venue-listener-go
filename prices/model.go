package prices

import (
	"math"
)

type PriceSummary struct {
	Assets []AssetPrice
}

type AssetPrice struct {
	Name      string  `json:"name"`
	Ticker    string  `json:"productId"`
	HighOffer float64 `json:"highOffer"`
	LowBid    float64 `json:"lowBid"`
	Spread    float64 `json:"spread"`
}

func (ap AssetPrice) NotSet() bool {
	if math.IsNaN(ap.HighOffer) || math.IsNaN(ap.LowBid) || math.IsNaN(ap.Spread) {
		return true
	}
	return false
}
