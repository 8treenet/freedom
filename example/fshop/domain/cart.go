package domain

import (
	"encoding/json"

	"github.com/8treenet/freedom/example/fshop/domain/aggregate"
	"github.com/8treenet/freedom/example/fshop/domain/dependency"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//绑定创建领域服务函数到框架，框架会根据客户的使用做依赖倒置和依赖注入的处理。
		initiator.BindService(func() *CartService {
			return &CartService{} //创建Cart领域服务
		})
		//控制器客户需要明确使用 InjectController
		initiator.InjectController(func(ctx freedom.Context) (service *CartService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

// CartService 领域服务.
type CartService struct {
	Worker      freedom.Worker         //运行时，一个请求绑定一个运行时
	CartRepo    dependency.CartRepo    //依赖倒置购物车资源库
	CartFactory *aggregate.CartFactory //依赖注入购物车聚合根工厂
	ShopFactory *aggregate.ShopFactory //依赖注入购买聚合根工厂
}

// Add 购物车增加商品
func (c *CartService) Add(userID, goodsID, goodsNum int) (e error) {
	//创建购物车增加商品聚合根
	cmd, e := c.CartFactory.NewCartAddCmd(goodsID, userID)
	if e != nil {
		return
	}
	return cmd.Run(goodsNum)
}

// Items 购物车全部商品项
func (c *CartService) Items(userID int) (json.Marshaler, error) {
	//创建购物车查询聚合根
	query, e := c.CartFactory.NewCartItemQuery(userID)
	if e != nil {
		return nil, e
	}
	return query, e
}

// DeleteAll 清空购物车
func (c *CartService) DeleteAll(userID int) (e error) {
	return c.CartRepo.DeleteAll(userID)
}

// Shop 购物车全部购买
func (c *CartService) Shop(userID int) (e error) {
	//使用抽象工厂 创建购物车类型
	shopType := c.ShopFactory.NewCartShopType()
	//使用抽象工厂 创建抽象聚合根
	cmd, e := c.ShopFactory.NewShopCmd(userID, shopType)
	if e != nil {
		return
	}
	return cmd.Shop()
}
