package aggregate

import (
	"errors"

	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
)

// CartAddCmd 添加购物车聚合根
type CartAddCmd struct {
	entity.User
	goods    entity.Goods
	cartRepo dependency.CartRepo
}

// Run .
func (cmd *CartAddCmd) Run(goodsNum int) error {
	if goodsNum > cmd.goods.Stock {
		return errors.New("the inventory is not enough for the supply")
	}

	if cartEntity, err := cmd.cartRepo.FindByGoodsID(cmd.User.ID, cmd.goods.ID); err == nil {
		//如果已在购物车找到商品，增加数量
		cartEntity.AddNum(goodsNum)
		return cmd.cartRepo.Save(cartEntity)
	}

	//创建商品到购物车
	_, e := cmd.cartRepo.New(cmd.User.ID, cmd.goods.ID, goodsNum)
	return e
}
