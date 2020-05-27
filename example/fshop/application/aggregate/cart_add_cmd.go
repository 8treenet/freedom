package aggregate

import (
	"errors"

	"github.com/8treenet/freedom/example/fshop/adapter/repository"
	"github.com/8treenet/freedom/example/fshop/application/entity"
)

// NewCartAddCmd 创建添加购物车聚合根，传入相关仓库的接口
func NewCartAddCmd(userRepo repository.UserRepo, cartRepo repository.CartRepo, goodsRepo repository.GoodsRepo) *CartAddCmd {
	return &CartAddCmd{
		userRepo:  userRepo,
		cartRepo:  cartRepo,
		goodsRepo: goodsRepo,
	}
}

// 添加购物车聚合根
type CartAddCmd struct {
	entity.User
	goods     entity.Goods
	userRepo  repository.UserRepo
	cartRepo  repository.CartRepo
	goodsRepo repository.GoodsRepo
}

// LoadEntity 加载依赖实体
func (cmd *CartAddCmd) LoadEntity(goodsId, userId int) error {
	user, e := cmd.userRepo.Get(userId)
	if e != nil {
		cmd.GetWorker().Logger().Error(e, "userId", userId)
		//用户不存在
		return e
	}
	cmd.User = *user

	goods, e := cmd.goodsRepo.Get(goodsId)
	if e != nil {
		//商品不存在
		cmd.GetWorker().Logger().Error(e, "userId", userId, "goodsId", goodsId)
		return e
	}
	cmd.goods = *goods
	return nil
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
