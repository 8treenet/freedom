package controllers

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/application"
	"github.com/8treenet/freedom/example/infra-example/infra"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/shop", &ShopController{})
	})
}

type ShopController struct {
	Runtime  freedom.Runtime
	OrderSev *application.OrderService
}

// GetBy handles the POST: /route 购买商品.
func (c *ShopController) Post() freedom.Result {
	var request struct {
		GoodsID int `json:"goodsId"` //商品id
		Num     int `json:"num"`     //购买数量
		UserID  int `json:"userId"`  //用户id
	}
	if e := c.Runtime.Ctx().ReadJSON(&request); e != nil {
		return &infra.JSONResponse{Error: e}
	}

	resp, e := c.OrderSev.Add(request.GoodsID, request.Num, request.UserID)
	if e != nil {
		return &infra.JSONResponse{Error: e}
	}
	return &infra.JSONResponse{Object: resp}
}
