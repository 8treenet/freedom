package controllers

import (
	"encoding/json"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/event-example/application/objects"
	"github.com/8treenet/freedom/infra/kafka"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/shop", &ShopController{})
	})
}

type ShopController struct {
	Runtime  freedom.Runtime
	Producer kafka.Producer
}

// Get handles the GET: /shop/:id route.
func (s *ShopController) GetBy(id int) string {
	data, _ := json.Marshal(objects.Goods{
		ID:     id,
		Amount: 10,
	})
	msg := s.Producer.NewMsg(EventSell, data)
	msg.SetHeaders(map[string]string{"x-action": "购买"}).SetRuntime(s.Runtime).Publish()
	return "ok"
}
