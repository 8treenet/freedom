package aggregate

import (
	"errors"
	"time"

	"github.com/8treenet/freedom/example/fshop/adapter/po"
	"github.com/8treenet/freedom/example/fshop/adapter/repository"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/infra/transaction"
)

// NewShopGoodsCmd 创建购买商品聚合根，传入相关仓库的接口
func NewShopGoodsCmd(userRepo repository.UserRepo, orderRepo repository.OrderRepo, goodsRepo repository.GoodsRepo, tx transaction.Transaction) *ShopGoodsCmd {
	return &ShopGoodsCmd{
		userRepo:  userRepo,
		orderRepo: orderRepo,
		goodsRepo: goodsRepo,
		tx:        tx,
	}
}

// 购买商品聚合根
type ShopGoodsCmd struct {
	entity.Order
	userRepo  repository.UserRepo
	orderRepo repository.OrderRepo
	goodsRepo repository.GoodsRepo
	tx        transaction.Transaction

	userEntity  entity.User
	goodsEntity *entity.Goods
}

// LoadEntity 加载依赖实体
func (cmd *ShopGoodsCmd) LoadEntity(goodsId, userId int) error {
	user, e := cmd.userRepo.Get(userId)
	if e != nil {
		cmd.GetWorker().Logger().Error(e, "userId", userId)
		//用户不存在
		return e
	}
	cmd.userEntity = *user

	cmd.goodsEntity, e = cmd.goodsRepo.Get(goodsId)
	if e != nil {
		return e
	}

	if order, e := cmd.orderRepo.New(); e != nil {
		return e
	} else {
		cmd.Order = *order
	}

	return nil
}

// Shop 购买
func (cmd *ShopGoodsCmd) Shop(goodsNum int) error {
	if goodsNum > cmd.goodsEntity.Stock {
		return errors.New("库存不足")
	}
	//扣库存
	cmd.goodsEntity.AddStock(-goodsNum)

	//设置订单总价格
	totalPrice := cmd.goodsEntity.Price * goodsNum
	cmd.SetTotalPrice(totalPrice)
	//设置订单的用户
	cmd.SetUserId(cmd.userEntity.Id)
	//增加订单的商品详情
	cmd.AddOrderDetal(&po.OrderDetail{OrderNo: cmd.OrderNo, GoodsId: cmd.goodsEntity.Id, GoodsName: cmd.goodsEntity.Name, Num: goodsNum, Created: time.Now(), Updated: time.Now()})

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
