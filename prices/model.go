package prices

import "time"

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
	Name      string
	Ticker    string
	HighOffer float64
	LowBid    float64
	Spread    float64
}
