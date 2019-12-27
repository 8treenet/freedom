package controllers

import (
	"encoding/json"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/msg-example/models"
	"github.com/8treenet/freedom/infra/kafka"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindController("/shop", &ShopController{})
	})
}

type ShopController struct {
	Runtime  freedom.Runtime
	Producer *kafka.Producer
}

// Get handles the GET: /shop/:id route.
func (s *ShopController) GetBy(id int) string {
	data, _ := json.Marshal(models.Goods{
		ID:     id,
		Amount: 10,
	})
	msg := s.Producer.NewMsg("event-sell", data)
	msg.SetHeaders(map[string]string{"Action": "购买"}).SetRuntime(s.Runtime).Send()
	return "ok"
}
