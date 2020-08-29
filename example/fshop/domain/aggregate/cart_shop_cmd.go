package aggregate

import (
	"errors"
	"time"

	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/example/fshop/domain/po"
	"github.com/8treenet/freedom/infra/transaction"
)

//CartShopCmd 购买商品聚合根
type CartShopCmd struct {
	entity.Order
	orderRepo dependency.OrderRepo
	goodsRepo dependency.GoodsRepo
	cartRepo  dependency.CartRepo
	tx        transaction.Transaction

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
		cmd.AddOrderDetal(&po.OrderDetail{OrderNo: cmd.OrderNo, GoodsID: goodsEntity.ID, GoodsName: goodsEntity.Name, Num: cmd.allCartEntity[i].Num, Created: time.Now(), Updated: time.Now()})
	}

	//设置订单总价格
	cmd.SetTotalPrice(totalPrice)
	//设置订单的用户
	cmd.SetUserID(cmd.userEntity.ID)

	//事务执行 创建 订单表、订单详情表，修改商品表的库存
	e = cmd.tx.Execute(func() error {
		for _, goodsEntity := range cmd.goodsEntityMap {
			if e := cmd.goodsRepo.Save(goodsEntity); e != nil {
				return e
			}
		}
		//清空购物车
		cmd.cartRepo.DeleteAll(cmd.UserID)

		//创建订单
		return cmd.orderRepo.Save(&cmd.Order)
	})

	if e != nil {
		return e
	}
	for _, goodsEntity := range cmd.goodsEntityMap {
		//发布领域事件，该商品被下单
		//需要配置 server/conf/infra/kafka.toml 生产者相关配置
		goodsEntity.DomainEvent("goods-shop", goodsEntity.ID)
	}
	return e
}
