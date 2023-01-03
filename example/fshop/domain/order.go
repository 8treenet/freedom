package domain

import (
	"github.com/8treenet/freedom/example/fshop/domain/aggregate"
	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/vo"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//绑定创建领域服务函数到框架，框架会根据客户的使用做依赖倒置和依赖注入的处理。
		initiator.BindService(func() *OrderService {
			return &OrderService{} //创建Order领域服务
		})
		//控制器客户需要明确使用 InjectController
		initiator.InjectController(func(ctx freedom.Context) (service *OrderService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

// OrderService 订单领域服务.
type OrderService struct {
	Worker       freedom.Worker          //运行时，一个请求绑定一个运行时
	OrderRepo    dependency.OrderRepo    //依赖倒置订单资源库
	OrderFactory *aggregate.OrderFactory //依赖注入订单工厂
}

// Pay 订单支付 .
func (o *OrderService) Pay(orderNo string, userID int) (e error) {
	cmd, e := o.OrderFactory.NewOrderPayCmd(orderNo, userID)
	if e != nil {
		return
	}
	return cmd.Pay()
}

// Items 订单列表.
func (o *OrderService) Items(userID int, page, pageSize int) (result []vo.OrderItemRes, totalPage int, e error) {
	items, totalPage, e := o.OrderRepo.Finds(userID, page, pageSize)
	if e != nil {
		return
	}
	for i := 0; i < len(items); i++ {
		item := vo.OrderItemRes{
			OrderNo:    items[i].OrderNo,
			TotalPrice: items[i].TotalPrice,
			Status:     items[i].Status,
		}
		for j := 0; j < len(items[i].Details); j++ {
			goodsItem := struct {
				GoodsID   int    // 商品id
				Num       int    // 数量
				GoodsName string // 商品名称
			}{
				items[i].Details[j].GoodsID,
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
func (o *OrderService) Delivery(req vo.DeliveryReq) (e error) {
	//创建订单发货聚合根
	cmd, e := o.OrderFactory.NewOrderDeliveryCmd(req.OrderNo, req.AdminID)
	if e != nil {
		return e
	}
	//传入快递单号执行命令
	return cmd.Run(req.TrackingNumber)
}
