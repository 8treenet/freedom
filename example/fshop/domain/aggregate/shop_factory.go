package aggregate

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/infra/transaction"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//绑定创建工厂函数到框架，框架会根据客户的使用做依赖倒置和依赖注入的处理。
		initiator.BindFactory(func() *ShopFactory {
			return &ShopFactory{} //创建ShopFactory
		})
	})
}

const (
	shopGoodsType = 1 //直接购买类型
	shopCartType  = 2 //购物车购买类型
)

// ShopFactory 购买聚合根抽象工厂
type ShopFactory struct {
	UserRepo  dependency.UserRepo     //依赖倒置用户资源库
	CartRepo  dependency.CartRepo     //依赖倒置购物车资源库
	GoodsRepo dependency.GoodsRepo    //依赖倒置商品资源库
	OrderRepo dependency.OrderRepo    //依赖倒置订单资源库
	TX        transaction.Transaction //依赖倒置事务组件
}

// NewGoodsShopType 创建商品购买类型
func (factory *ShopFactory) NewGoodsShopType(goodsID, goodsNum int) ShopType {
	return &shopType{
		stype:    shopGoodsType,
		goodsID:  goodsID,
		goodsNum: goodsNum,
	}
}

// NewCartShopType 创建购物车购买类型
func (factory *ShopFactory) NewCartShopType() ShopType {
	return &shopType{
		stype: shopCartType,
	}
}

// NewShopCmd 创建抽象聚合根
func (factory *ShopFactory) NewShopCmd(userID int, stype ShopType) (ShopCmd, error) {
	if stype.GetType() == 2 {
		return factory.newCartShopCmd(userID)
	}
	goodsID, goodsNum := stype.GetDirectGoods()
	return factory.newGoodsShopCmd(userID, goodsID, goodsNum)
}

// newGoodsShopCmd 创建购买商品聚合根
func (factory *ShopFactory) newGoodsShopCmd(userID, goodsID, goodsNum int) (*GoodsShopCmd, error) {
	user, e := factory.UserRepo.Get(userID)
	if e != nil {
		//用户不存在
		return nil, e
	}
	order, e := factory.OrderRepo.New()
	if e != nil {
		return nil, e
	}

	goodsEntity, e := factory.GoodsRepo.Get(goodsID)
	if e != nil {
		return nil, e
	}
	cmd := &GoodsShopCmd{
		Order:       *order,
		goodsNum:    goodsNum,
		userEntity:  user,
		goodsEntity: goodsEntity,
		orderRepo:   factory.OrderRepo,
		goodsRepo:   factory.GoodsRepo,
		tx:          factory.TX,
	}
	return cmd, nil
}

// newCartShopCmd 创建购买聚合根
func (factory *ShopFactory) newCartShopCmd(userID int) (*CartShopCmd, error) {
	user, e := factory.UserRepo.Get(userID)
	if e != nil {
		user.GetWorker().Logger().Error(e, "userId", userID)
		//用户不存在
		return nil, e
	}
	cmd := &CartShopCmd{
		orderRepo: factory.OrderRepo,
		goodsRepo: factory.GoodsRepo,
		cartRepo:  factory.CartRepo,
		tx:        factory.TX,
	}
	cmd.userEntity = *user
	cmd.allCartEntity, e = factory.CartRepo.FindAll(user.ID)
	if e != nil {
		return nil, e
	}

	cmd.goodsEntityMap = make(map[int]*entity.Goods)
	for i := 0; i < len(cmd.allCartEntity); i++ {
		goodsEntity, e := factory.GoodsRepo.Get(cmd.allCartEntity[i].GoodsID)
		if e != nil {
			return nil, e
		}
		cmd.goodsEntityMap[goodsEntity.ID] = goodsEntity
	}
	return cmd, nil
}

type shopType struct {
	stype    int
	goodsID  int
	goodsNum int
}

func (st *shopType) GetType() int {
	return st.stype
}

func (st *shopType) GetDirectGoods() (int, int) {
	return st.goodsID, st.goodsNum
}
