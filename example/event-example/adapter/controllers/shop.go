package controllers

import (
	"encoding/json"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/event-example/adapter/dto"
	"github.com/8treenet/freedom/infra/kafka"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/shop", &ShopController{})
	})
}

type ShopController struct {
	Worker   freedom.Worker
	Producer kafka.Producer
}

// Get handles the GET: /shop/:id route.
func (s *ShopController) GetBy(id int) string {
	data, _ := json.Marshal(dto.Goods{
		ID:     id,
		Amount: 10,
	})
	msg := s.Producer.NewMsg(EventSell, data)
	msg.SetHeaders(map[string]string{"x-action": "购买"}).SetWorker(s.Worker).Publish()
	return "ok"
}
