package aggregate

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/adapter/repository"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindFactory(func() *CartFactory {
			return &CartFactory{}
		})
	})
}

// CartFactory 购物车聚合根工厂
type CartFactory struct {
	UserRepo  repository.UserRepo  //依赖倒置用户资源库
	CartRepo  repository.CartRepo  //依赖倒置购物车资源库
	GoodsRepo repository.GoodsRepo //依赖倒置商品资源库
	OrderRepo repository.OrderRepo //依赖倒置订单资源库
}

// NewCartAddCmd 创建添加聚合根
func (factory *CartFactory) NewCartAddCmd(goodsId, userId int) (*CartAddCmd, error) {
	user, e := factory.UserRepo.Get(userId)
	if e != nil {
		user.GetWorker().Logger().Error(e, "userId", userId)
		//用户不存在
		return nil, e
	}

	goods, e := factory.GoodsRepo.Get(goodsId)
	if e != nil {
		//商品不存在
		goods.GetWorker().Logger().Error(e, "userId", userId, "goodsId", goodsId)
		return nil, e
	}
	cmd := &CartAddCmd{
		cartRepo: factory.CartRepo,
	}
	cmd.User = *user
	cmd.goods = *goods
	return cmd, nil
}

// NewCartItemQuery 创建查询聚合根
func (factory *CartFactory) NewCartItemQuery(userId int) (*CartItemQuery, error) {
	user, e := factory.UserRepo.Get(userId)
	if e != nil {
		user.GetWorker().Logger().Error(e, "userId", userId)
		//用户不存在
		return nil, e
	}

	query := &CartItemQuery{}
	query.User = *user
	query.goodsMap = make(map[int]*entity.Goods)

	query.allCart, e = factory.CartRepo.FindAll(query.User.Id)
	for i := 0; i < len(query.allCart); i++ {
		goodsEntity, e := factory.GoodsRepo.Get(query.allCart[i].GoodsId)
		if e != nil {
			return nil, e
		}
		query.goodsMap[goodsEntity.Id] = goodsEntity
	}
	return query, nil
}
