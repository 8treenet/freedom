package aggregate

import (
	"errors"
	"time"

	"github.com/8treenet/freedom/example/fshop/adapter/repository"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/example/fshop/domain/po"
	"github.com/8treenet/freedom/infra/transaction"
)

// 购买商品聚合根
type GoodsShopCmd struct {
	entity.Order
	userEntity  *entity.User
	goodsEntity *entity.Goods

	orderRepo repository.OrderRepo
	goodsRepo repository.GoodsRepo
	tx        transaction.Transaction
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
	cmd.SetUserId(cmd.userEntity.Id)
	//增加订单的商品详情
	cmd.AddOrderDetal(&po.OrderDetail{OrderNo: cmd.OrderNo, GoodsId: cmd.goodsEntity.Id, GoodsName: cmd.goodsEntity.Name, Num: cmd.goodsNum, Created: time.Now(), Updated: time.Now()})

	//事务执行 创建 订单表、订单详情表，修改商品表的库存
	e := cmd.tx.Execute(func() error {
		if e := cmd.orderRepo.Save(&cmd.Order); e != nil {
			return e
		}
		if e := cmd.goodsRepo.Save(cmd.goodsEntity); e != nil {
			return e
		}
		return nil
	})

	if e == nil {
		//发布领域事件，该商品被下单
		//需要配置 server/conf/infra/kafka.toml 生产者相关配置
		cmd.goodsEntity.DomainEvent("goods-shop", cmd.goodsEntity.Id)
	}

	return e
}
