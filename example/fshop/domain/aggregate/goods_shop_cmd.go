package aggregate

import (
	"errors"
	"time"

	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/example/fshop/domain/event"
	"github.com/8treenet/freedom/example/fshop/domain/po"
	"github.com/8treenet/freedom/example/fshop/infra/domainevent"
)

//GoodsShopCmd 购买商品聚合根
type GoodsShopCmd struct {
	entity.Order
	userEntity  *entity.User
	goodsEntity *entity.Goods

	orderRepo dependency.OrderRepo
	goodsRepo dependency.GoodsRepo
	tx        *domainevent.EventTransaction
	goodsNum  int
}

// Shop 购买
func (cmd *GoodsShopCmd) Shop() error {
	if cmd.goodsNum > cmd.goodsEntity.Stock {
		return errors.New("库存不足")
	}
	//扣库存
	cmd.goodsEntity.AddStock(-cmd.goodsNum)

	//设置订单总价格
	totalPrice := cmd.goodsEntity.Price * cmd.goodsNum
	cmd.SetTotalPrice(totalPrice)
	//设置订单的用户
	cmd.SetUserID(cmd.userEntity.ID)
	//增加订单的商品详情
	cmd.AddOrderDetail(&po.OrderDetail{OrderNo: cmd.OrderNo, GoodsID: cmd.goodsEntity.ID, GoodsName: cmd.goodsEntity.Name, Num: cmd.goodsNum, Created: time.Now(), Updated: time.Now()})

	//订单实体加入购买事件
	cmd.Order.AddPubEvent(&event.ShopGoods{
		UserID:    cmd.UserID,
		OrderNO:   cmd.OrderNo,
		GoodsID:   cmd.goodsEntity.ID,
		GoodsNum:  cmd.goodsNum,
		GoodsName: cmd.goodsEntity.Name,
	})

	//使用事务组件保证一致性 1.修改商品库存, 2.创建订单, 3.事件表增加记录
	//Execute 如果返回错误 会触发回滚。成功会调用infra/domainevent/EventManager.push
	e := cmd.tx.Execute(func() error {
		if e := cmd.goodsRepo.Save(cmd.goodsEntity); e != nil {
			return e
		}
		return cmd.orderRepo.Save(&cmd.Order)
	})
	return e
}
