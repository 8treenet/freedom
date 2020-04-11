package controllers

import (
	"github.com/8treenet/freedom/example/infra-example/application"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/infra"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/goods", &GoodsController{})
	})
}

type GoodsController struct {
	Runtime     freedom.Runtime
	GoodsSev    *application.GoodsService
	JSONRequest *infra.JSONRequest
}

// GetBy handles the GET: /goods/:id route 获取指定商品.
func (goods *GoodsController) GetBy(goodsID int) freedom.Result {
	obj, err := goods.GoodsSev.Get(goodsID)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: obj}
}

// GetBy handles the GET: /goods/:id route 获取商品列表.
func (goods *GoodsController) Get() freedom.Result {
	objs, err := goods.GoodsSev.GetAll()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: objs}
}

// GetBy handles the PUT: /goods/stock route 增加商品库存.
func (goods *GoodsController) PutStock() freedom.Result {
	var request struct {
		GoodsID int `json:"goodsId" validate:"required"` //商品id
		Num     int `validate:"min=1,max=15"`
	}
	//使用自定义的json基础设施
	if e := goods.JSONRequest.ReadBodyJSON(&request); e != nil {
		return &infra.JSONResponse{Error: e}
	}

	if e := goods.GoodsSev.AddStock(request.GoodsID, request.Num); e != nil {
		return &infra.JSONResponse{Error: e}
	}
	return &infra.JSONResponse{}
}
