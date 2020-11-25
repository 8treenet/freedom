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

//CartShopCmd 购买商品聚合根
type CartShopCmd struct {
	entity.Order
	orderRepo dependency.OrderRepo
	goodsRepo dependency.GoodsRepo
	cartRepo  dependency.CartRepo
	tx        *domainevent.EventTransaction

	userEntity     entity.User
	allCartEntity  []*entity.Cart
	goodsEntityMap map[int]*entity.Goods
}

// Shop 购买
func (cmd *CartShopCmd) Shop() error {
	order, e := cmd.orderRepo.New()
	if e != nil {
		return e
	}
	cmd.Order = *order

	var totalPrice int

	for i := 0; i < len(cmd.allCartEntity); i++ {
		goodsEntity := cmd.goodsEntityMap[cmd.allCartEntity[i].GoodsID]
		//判断 购物车商品库存是否不足
		if cmd.allCartEntity[i].Num > goodsEntity.Stock {
			return errors.New("库存不足")
		}

		//扣库存
		goodsEntity.AddStock(-cmd.allCartEntity[i].Num)
		totalPrice += goodsEntity.Price * cmd.allCartEntity[i].Num

		//增加订单的商品详情
		cmd.AddOrderDetail(&po.OrderDetail{OrderNo: cmd.OrderNo, GoodsID: goodsEntity.ID, GoodsName: goodsEntity.Name, Num: cmd.allCartEntity[i].Num, Created: time.Now(), Updated: time.Now()})

		//订单实体加入购买事件
		cmd.Order.AddPubEvent(&event.ShopGoods{
			UserID:    cmd.userEntity.ID,
			OrderNO:   cmd.OrderNo,
			GoodsID:   goodsEntity.ID,
			GoodsNum:  cmd.allCartEntity[i].Num,
			GoodsName: goodsEntity.Name,
		})
	}

	//设置订单总价格
	cmd.SetTotalPrice(totalPrice)
	//设置订单的用户
	cmd.SetUserID(cmd.userEntity.ID)

	//使用事务组件保证一致性 1.修改商品库存, 2.清空购物车, 3.创建订单, 4.事件表增加记录
	//Execute 如果返回错误 会触发回滚。成功会调用infra/domainevent/EventManager.push
	e = cmd.tx.Execute(func() error {
		for _, goodsEntity := range cmd.goodsEntityMap {
			if e := cmd.goodsRepo.Save(goodsEntity); e != nil {
				return e
			}
		}

		//清空购物车
		if err := cmd.cartRepo.DeleteAll(cmd.UserID); err != nil {
			return err
		}

		//创建订单
		return cmd.orderRepo.Save(&cmd.Order)
	})
	return e
}
