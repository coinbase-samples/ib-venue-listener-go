package order

import "time"

type Update struct {
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
