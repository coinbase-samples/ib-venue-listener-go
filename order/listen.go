package order

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coinbase-samples/ib-venue-listener-go/cloud"
	"github.com/coinbase-samples/ib-venue-listener-go/model"

	"github.com/coinbase-samples/ib-venue-listener-go/config"
	log "github.com/sirupsen/logrus"
)

func ProcessOrderMessage(app config.AppConfig, message []byte) error {
	var ud = &model.OrderUpdate{}
	if err := json.Unmarshal(message, ud); err != nil {
		return fmt.Errorf("unable to umarshal json: %s - msg: %v", string(message), err)
	}

	log.Debugf("parsed order message - %v", ud)

	writeOrderUpdatesToQueue(app, ud)

	return nil
}

func writeOrderUpdatesToQueue(
	app config.AppConfig,
	orderUpdate *model.OrderUpdate,
) {
	// loop in loop because everything is an array for some reason
	for _, event := range orderUpdate.Events {

		if event.Type == "error" {
			if len(event.Message) > 0 {
				log.Errorf("Orders channel error: %s", event.Message)
			}
			continue
		}

		for _, order := range event.Orders {

			message := model.ConvertOrderFillMessage(order, orderUpdate.Timestamp)

			val, err := json.Marshal(message)

			if err != nil {
				log.Errorf("error marshalling order fill message", err)
			}
			log.Warnf("publishing order feed update - %s", val)
			if err := cloud.SqsSendMessage(
				context.Background(),
				app,
				app.OrderFillQueueUrl,
				order.ClientOrderID,
				string(val),
			); err != nil {
				log.Errorf("unable to send SQS message: %v", err)
			}

		}
	}
}
