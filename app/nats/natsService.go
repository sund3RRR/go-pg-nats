package nats

import (
	"app/db"
	"app/order"
	"encoding/json"
	"fmt"

	"github.com/nats-io/stan.go"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

type NatsService struct {
	Logger    *zap.Logger
	DBService *db.DatabaseService
	Cache     *cache.Cache
}

func (nats *NatsService) HandleMessage(msg *stan.Msg) {
	nats.Logger.Info(fmt.Sprintf("Received nats message. Message CRC32:%d", msg.CRC32))

	var data order.Order
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		nats.Logger.Error("Received bad json data from channel", zap.Error(err))
		return
	}

	data.Delivery.OrderUID = data.OrderUID
	data.Payment.OrderUID = data.OrderUID

	for i := 0; i < len(data.Items); i++ {
		data.Items[i].OrderUID = data.OrderUID
	}

	nats.Cache.Set(data.OrderUID, data, cache.NoExpiration)
	go nats.DBService.DumpCachedOrder(data.OrderUID)
}
