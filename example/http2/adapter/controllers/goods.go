package controllers

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/http2/adapter/dto"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/goods", &GoodsController{})
	})
}

type GoodsController struct {
	Worker freedom.Worker
}

// Get handles the GET: /goods/:id route.
func (goods *GoodsController) GetBy(id int) (result dto.Goods, e error) {
	goods.Worker.Logger().Info("我是GoodsController.GetByID控制器,返回商品名称和价格")
	//打印出bus的数据
	goods.Worker.Logger().Error("bus", goods.Worker.Bus())
	switch id {
	case 1:
		result.Name = "冈本"
		result.Price = 18
	case 2:
		result.Name = "杰士邦"
		result.Price = 15
	case 3:
		result.Name = "杜蕾斯"
		result.Price = 20
	case 4:
		result.Name = "第六感"
		result.Price = 10
	default:
		result.Name = "未知商品"
		result.Price = 0
	}
	return
}
