package aggregate

import (
	"errors"

	"github.com/8treenet/freedom/example/fshop/adapter/repository"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
)

// 添加购物车聚合根
type CartAddCmd struct {
	entity.User
	goods    entity.Goods
	cartRepo repository.CartRepo
}

func (cmd *CartAddCmd) Run(goodsNum int) error {
	if goodsNum > cmd.goods.Stock {
		return errors.New("the inventory is not enough for the supply")
	}

	if cartEntity, err := cmd.cartRepo.FindByGoodsId(cmd.User.Id, cmd.goods.Id); err == nil {
		//如果已在购物车找到商品，增加数量
		cartEntity.AddNum(goodsNum)
		return cmd.cartRepo.Save(cartEntity)
	}

	//创建商品到购物车
	_, e := cmd.cartRepo.New(cmd.User.Id, cmd.goods.Id, goodsNum)
	return e
}
