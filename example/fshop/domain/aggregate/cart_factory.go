package aggregate

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//绑定创建工厂函数到框架，框架会根据客户的使用做依赖倒置和依赖注入的处理。
		initiator.BindFactory(func() *CartFactory {
			return &CartFactory{} //创建CartFactory
		})
	})
}

// CartFactory 购物车聚合根工厂
type CartFactory struct {
	UserRepo  dependency.UserRepo  //依赖倒置用户资源库
	CartRepo  dependency.CartRepo  //依赖倒置购物车资源库
	GoodsRepo dependency.GoodsRepo //依赖倒置商品资源库
	OrderRepo dependency.OrderRepo //依赖倒置订单资源库
}

// NewCartAddCmd 创建添加聚合根
func (factory *CartFactory) NewCartAddCmd(goodsID, userID int) (*CartAddCmd, error) {
	user, e := factory.UserRepo.Get(userID)
	if e != nil {
		user.Worker().Logger().Error(e, "userId", userID)
		//用户不存在
		return nil, e
	}

	goods, e := factory.GoodsRepo.Get(goodsID)
	if e != nil {
		//商品不存在
		goods.Worker().Logger().Error(e, "userId", userID, "goodsId", goodsID)
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
func (factory *CartFactory) NewCartItemQuery(userID int) (*CartItemQuery, error) {
	user, e := factory.UserRepo.Get(userID)
	if e != nil {
		user.Worker().Logger().Error(e, "userId", userID)
		//用户不存在
		return nil, e
	}

	query := &CartItemQuery{}
	query.User = *user
	query.goodsMap = make(map[int]*entity.Goods)

	query.allCart, e = factory.CartRepo.FindAll(query.User.ID)
	for i := 0; i < len(query.allCart); i++ {
		goodsEntity, e := factory.GoodsRepo.Get(query.allCart[i].GoodsID)
		if e != nil {
			return nil, e
		}
		query.goodsMap[goodsEntity.ID] = goodsEntity
	}
	return query, nil
}
