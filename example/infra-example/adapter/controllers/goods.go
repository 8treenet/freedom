package controllers

import (
	"github.com/8treenet/freedom/example/infra-example/domain"
	"github.com/8treenet/freedom/example/infra-example/domain/event"
	"github.com/8treenet/freedom/example/infra-example/infra/domainevent"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/infra"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/goods", &GoodsController{})
		initiator.ListenEvent((&event.ShopGoods{}).Topic(), "GoodsController.PostStock")
	})
}

// GoodsController .
type GoodsController struct {
	Worker       freedom.Worker
	GoodsSev     *domain.GoodsService
	Request      *infra.Request
	EventManager *domainevent.EventManager //领域事件组件
}

// GetBy handles the GET: /goods/:id route 获取指定商品.
func (goods *GoodsController) GetBy(goodsID int) freedom.Result {
	obj, err := goods.GoodsSev.Get(goodsID)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: obj}
}

// Get handles the GET: /goods/:id route 获取商品列表.
func (goods *GoodsController) Get() freedom.Result {
	objs, err := goods.GoodsSev.GetAll()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: objs}
}

// PostStock handles the POST: /goods/stock route 增加商品库存.
func (goods *GoodsController) PostStock() freedom.Result {
	shopEvent := &event.ShopGoods{}
	//使用自定义的json组件
	if e := goods.Request.ReadJSON(shopEvent); e != nil {
		return &infra.JSONResponse{Error: e}
	}
	goods.Worker.Logger().Infof("领域事件消费 Topic:%s, %+v", shopEvent.Topic(), shopEvent)

	//先插入到事件表
	if e := goods.EventManager.InsertSubEvent(shopEvent); e != nil {
		return &infra.JSONResponse{Error: e}
	}

	if e := goods.GoodsSev.ShopEvent(shopEvent); e != nil {
		return &infra.JSONResponse{Error: e}
	}
	return &infra.JSONResponse{}
}
