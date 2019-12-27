package controllers

import (
	"github.com/8treenet/freedom/example/infra-example/business/services"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/infra"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindController("/goods", &GoodsController{})
	})
}

type GoodsController struct {
	Runtime  freedom.Runtime
	GoodsSev *services.GoodsService
}

// GetBy handles the GET: /goods/:id route 获取指定商品.
func (c *GoodsController) GetBy(goodsID int) freedom.Result {
	obj, err := c.GoodsSev.Get(goodsID)
	if err != nil {
		return &infra.Response{ErrCode: -1, ErrMsg: err.Error()}
	}
	return &infra.Response{Object: obj}
}

// GetBy handles the GET: /goods/:id route 获取商品列表.
func (c *GoodsController) Get() freedom.Result {
	objs, err := c.GoodsSev.GetAll()
	if err != nil {
		return &infra.Response{ErrCode: -1, ErrMsg: err.Error()}
	}

	return &infra.Response{Object: objs}
}

// GetBy handles the PUT: /goods/stock route 增加商品库存.
func (c *GoodsController) PutStock() freedom.Result {
	var request struct {
		GoodsID int `json:"goodsId"` //商品id
		Num     int `json:"num"`     //增加数量
	}
	if e := c.Runtime.Ctx().ReadJSON(&request); e != nil {
		return &infra.Response{ErrCode: -1, ErrMsg: e.Error()}
	}
	if e := c.GoodsSev.AddStock(request.GoodsID, request.Num); e != nil {
		return &infra.Response{ErrCode: -1, ErrMsg: e.Error()}
	}
	return &infra.Response{}
}
