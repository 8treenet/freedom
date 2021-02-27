package controller

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/common"
	"github.com/8treenet/freedom/example/infra-example/domain"
	"github.com/8treenet/freedom/example/infra-example/infra"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/shop", &ShopController{})
	})
}

// ShopController .
type ShopController struct {
	Worker   freedom.Worker
	OrderSev *domain.OrderService
	Request  *infra.Request
}

// Post handles the POST: /route 购买商品.
func (c *ShopController) Post() freedom.Result {
	var request struct {
		GoodsID int `json:"goodsId"` //商品id
		Num     int `json:"num"`     //购买数量
		UserID  int `json:"userId"`  //用户id
	}
	if e := c.Request.ReadJSON(&request); e != nil {
		return &infra.JSONResponse{Error: e}
	}

	var resErr error
	for i := 0; i < 3; i++ {
		resErr = c.OrderSev.Shop(request.GoodsID, request.Num, request.UserID)
		if resErr == common.ErrVersionExpired {
			continue //乐观锁重试
		}
		break
	}

	if resErr != nil {
		return &infra.JSONResponse{Error: resErr}
	}
	return &infra.JSONResponse{}
}
