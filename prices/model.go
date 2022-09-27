package prices

import (
	"math"
	"time"
)

type OrderBookUpdate struct {
	Channel     string    `json:"channel"`
	Timestamp   time.Time `json:"timestamp"`
	SequenceNum int       `json:"sequence_num"`
	Events      []struct {
		Type      string `json:"type"`
		ProductID string `json:"product_id"`
		Updates   []struct {
			Side      string    `json:"side"`
			EventTime time.Time `json:"event_time"`
			Px        string    `json:"px"`
			Qty       string    `json:"qty"`
		} `json:"updates"`
	} `json:"events"`
}

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
