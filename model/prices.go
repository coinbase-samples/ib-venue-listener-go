package model

import (
	"math"
	"time"
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

type AssetPriceUpdate struct {
	Name      string  `json:"name"`
	Ticker    string  `json:"productId"`
	HighOffer float64 `json:"highOffer"`
	LowBid    float64 `json:"lowBid"`
	Spread    float64 `json:"spread"`
}

type Asset struct {
	AssetId              string    `json:"assetId" dynamodbav:"productId"`
	Filter               string    `json:"filter" dynamodbav:"filter"`
	Ticker               string    `json:"ticker" dynamodbav:"ticker"`
	Name                 string    `json:"name" dynamodbav:"name"`
	MinTransactionAmount string    `json:"minTransactionAmount" dynamodbav:"minTransactionAmount"`
	MaxTransactionAmount string    `json:"maxTransactionAmount" dynamodbav:"maxTransactionAmount"`
	Slippage             string    `json:"slippage" dynamodbav:"slippage"`
	HighOffer            string    `json:"highOffer" dynamodbav:"highOffer"`
	LowBid               string    `json:"lowBid" dynamodbav:"lowBid"`
	Spread               string    `json:"spread" dynamodbav:"spread"`
	CreatedAt            time.Time `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt" dynamodbav:"updatedAt"`
	MarketCap            string    `json:"marketCap" dynamodbav:"marketCap"`
	Volume               string    `json:"volume" dynamodbav:"volume"`
	Supply               string    `json:"supply" dynamodbav:"supply"`
	Direction            string    `json:"direction" dynamodbav:"direction"`
}
