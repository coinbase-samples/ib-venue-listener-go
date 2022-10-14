package model

import "time"

type OrderBookUpdate struct {
	Channel     string    `json:"channel"`
	Timestamp   time.Time `json:"timestamp"`
	SequenceNum int       `json:"sequence_num"`
	Events      []struct {
		Type             string `json:"type"`
		HeartbeatCounter string `json:"heartbeat_counter"`
		Message          string `json:"message"`
		ProductID        string `json:"product_id"`
		Updates          []struct {
			Side      string    `json:"side"`
			EventTime time.Time `json:"event_time"`
			Px        string    `json:"px"`
			Qty       string    `json:"qty"`
		} `json:"updates"`
		Subscriptions struct {
			Orders []string `json:"orders"`
		} `json:"subscriptions"`
	} `json:"events"`
}

type OrderUpdateItem struct {
	OrderID       string `json:"order_id"`
	ClientOrderID string `json:"client_order_id"`
	CumQty        string `json:"cum_qty"`
	LeavesQty     string `json:"leaves_qty"`
	AvgPx         string `json:"avg_px"`
	Fees          string `json:"fees"`
	Status        string `json:"status"`
}

type OrderUpdate struct {
	Channel     string    `json:"channel"`
	Timestamp   time.Time `json:"timestamp"`
	SequenceNum int       `json:"sequence_num"`
	Events      []struct {
		Type             string            `json:"type"`
		HeartbeatCounter string            `json:"heartbeat_counter"`
		Message          string            `json:"message"`
		Orders           []OrderUpdateItem `json:"orders"`
		Subscriptions    struct {
			L2Data []string `json:"l2_data"`
		} `json:"subscriptions"`
	} `json:"events"`
}

type OrderFillMessage struct {
	OrderID       string    `json:"order_id"`
	ClientOrderID string    `json:"client_order_id"`
	CumQty        string    `json:"cum_qty"`
	LeavesQty     string    `json:"leaves_qty"`
	AvgPx         string    `json:"avg_px"`
	Fees          string    `json:"fees"`
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
}

func ConvertOrderFillMessage(o OrderUpdateItem, ts time.Time) OrderFillMessage {
	return OrderFillMessage{
		OrderID:       o.OrderID,
		ClientOrderID: o.ClientOrderID,
		CumQty:        o.CumQty,
		LeavesQty:     o.LeavesQty,
		AvgPx:         o.AvgPx,
		Fees:          o.Fees,
		Status:        o.Status,
		Timestamp:     ts,
	}
}
