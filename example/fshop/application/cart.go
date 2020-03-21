package application

import (
	"github.com/8treenet/freedom/example/fshop/adapter/repository"
	"github.com/8treenet/freedom/example/fshop/application/aggregate"
	"github.com/8treenet/freedom/example/fshop/application/dto"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindService(func() *Cart {
			return &Cart{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *Cart) {
			initiator.GetService(ctx, &service)
			return
		})
	})
}

// Cart 领域服务.
type Cart struct {
	Runtime   freedom.Runtime
	UserRepo  repository.UserRepo
	CartRepo  repository.CartRepo
	GoodsRepo repository.GoodsRepo
}

// Add 购物车增加商品
func (c *Cart) Add(userId, goodsId, goodsNum int) (e error) {
	cmd := aggregate.NewCartAddCmd(c.UserRepo, c.CartRepo, c.GoodsRepo)
	e = cmd.LoadEntity(goodsId, goodsNum)
	if e != nil {
		return
	}

	return cmd.Run(goodsNum)
}

// Items 购物车全部商品项
func (c *Cart) Items(userId int) (items dto.CartItemRes, e error) {
	//创建查询购物车上牌聚合根
	query := aggregate.NewCartItemQuery(c.UserRepo, c.CartRepo, c.GoodsRepo)
	e = query.LoadEntity(userId)
	if e != nil {
		return
	}

	//查询购物车全部商品
	if e = query.QueryAllItem(); e != nil {
		return
	}

	query.VisitAllItem(func(id, goodsId int, goodsName string, goodsNum, totalPrice int) {
		items.Items = append(items.Items, struct {
			Id         int
			GoodsId    int
			GoodsName  string
			GoodsNum   int
			TotalPrice int
		}{
			id,
			goodsId,
			goodsName,
			goodsNum,
			totalPrice,
		})
	})
	items.TotalPrice = query.AllItemTotalPrice()
	return
}

// DeleteAll 清空购物车
func (c *Cart) DeleteAll(userId int) (e error) {
	return c.CartRepo.DeleteAll(userId)
}
