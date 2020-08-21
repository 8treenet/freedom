package domain

import (
	"github.com/8treenet/freedom/example/fshop/domain/aggregate"
	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/dto"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//绑定创建领域服务函数到框架，框架会根据客户的使用做依赖倒置和依赖注入的处理。
		initiator.BindService(func() *Order {
			return &Order{} //创建Order领域服务
		})
		//控制器客户需要明确使用 InjectController
		initiator.InjectController(func(ctx freedom.Context) (service *Order) {
			initiator.GetService(ctx, &service)
			return
		})
	})
}

// Order 订单领域服务.
type Order struct {
	Worker       freedom.Worker          //运行时，一个请求绑定一个运行时
	OrderRepo    dependency.OrderRepo    //依赖倒置订单资源库
	OrderFactory *aggregate.OrderFactory //依赖注入订单工厂
}

// Pay 订单支付 .
func (o *Order) Pay(orderNo string, userId int) (e error) {
	cmd, e := o.OrderFactory.NewOrderPayCmd(orderNo, userId)
	if e != nil {
		return
	}
	return cmd.Pay()
}

// Items 订单列表.
func (o *Order) Items(userId int, page, pageSize int) (result []dto.OrderItemRes, totalPage int, e error) {
	items, totalPage, e := o.OrderRepo.Finds(userId, page, pageSize)
	if e != nil {
		return
	}
	for i := 0; i < len(items); i++ {
		item := dto.OrderItemRes{
			OrderNo:    items[i].OrderNo,
			TotalPrice: items[i].TotalPrice,
			Status:     items[i].Status,
		}
		for j := 0; j < len(items[i].Details); j++ {
			goodsItem := struct {
				GoodsId   int    // 商品id
				Num       int    // 数量
				GoodsName string // 商品名称
			}{
				items[i].Details[j].GoodsId,
				items[i].Details[j].Num,
				items[i].Details[j].GoodsName,
			}
			item.GoodsItems = append(item.GoodsItems, goodsItem)
		}

		result = append(result, item)
	}
	return
}

// Delivery 管理员发货服务
func (o *Order) Delivery(req dto.DeliveryReq) (e error) {
	//创建订单发货聚合根
	cmd, e := o.OrderFactory.NewOrderDeliveryCmd(req.OrderNo, req.AdminId)
	if e != nil {
		return e
	}
	//传入快递单号执行命令
	return cmd.Run(req.TrackingNumber)
}
