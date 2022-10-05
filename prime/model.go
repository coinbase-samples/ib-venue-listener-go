package prime

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

type OrderUpdate struct {
	Channel     string    `json:"channel"`
	Timestamp   time.Time `json:"timestamp"`
	SequenceNum int       `json:"sequence_num"`
	Events      []struct {
		Type   string `json:"type"`
		Orders []struct {
			OrderID       string `json:"order_id"`
			ClientOrderID string `json:"client_order_id"`
			CumQty        string `json:"cum_qty"`
			LeavesQty     string `json:"leaves_qty"`
			AvgPx         string `json:"avg_px"`
			Fees          string `json:"fees"`
			Status        string `json:"status"`
		} `json:"orders"`
	} `json:"events"`
}
