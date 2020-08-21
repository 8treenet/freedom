package controllers

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/http2/domain"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/shop", &ShopController{})
	})
}

type ShopController struct {
	Worker   freedom.Worker
	Shopping *domain.ShopService
}

// Get handles the GET: /shop/:id route.
func (s *ShopController) GetBy(id int) string {
	s.Worker.Logger().Info("我是控制器 ShopController.GetByID")
	return s.Shopping.Shopping(id)
}
