package aggregate

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/infra/transaction"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindFactory(func() *ShopFactory {
			return &ShopFactory{}
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
func (factory *ShopFactory) NewGoodsShopType(goodsId, goodsNum int) ShopType {
	return &shopType{
		stype:    shopGoodsType,
		goodsId:  goodsId,
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
func (factory *ShopFactory) NewShopCmd(userId int, stype ShopType) (ShopCmd, error) {
	if stype.GetType() == 2 {
		return factory.newCartShopCmd(userId)
	}
	goodsId, goodsNum := stype.GetDirectGoods()
	return factory.newGoodsShopCmd(userId, goodsId, goodsNum)
}

// newGoodsShopCmd 创建购买商品聚合根
func (factory *ShopFactory) newGoodsShopCmd(userId, goodsId, goodsNum int) (*GoodsShopCmd, error) {
	user, e := factory.UserRepo.Get(userId)
	if e != nil {
		//用户不存在
		return nil, e
	}
	order, e := factory.OrderRepo.New()
	if e != nil {
		return nil, e
	}

	goodsEntity, e := factory.GoodsRepo.Get(goodsId)
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
func (factory *ShopFactory) newCartShopCmd(userId int) (*CartShopCmd, error) {
	user, e := factory.UserRepo.Get(userId)
	if e != nil {
		user.GetWorker().Logger().Error(e, "userId", userId)
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
	cmd.allCartEntity, e = factory.CartRepo.FindAll(user.Id)
	if e != nil {
		return nil, e
	}

	cmd.goodsEntityMap = make(map[int]*entity.Goods)
	for i := 0; i < len(cmd.allCartEntity); i++ {
		goodsEntity, e := factory.GoodsRepo.Get(cmd.allCartEntity[i].GoodsId)
		if e != nil {
			return nil, e
		}
		cmd.goodsEntityMap[goodsEntity.Id] = goodsEntity
	}
	return cmd, nil
}

type shopType struct {
	stype    int
	goodsId  int
	goodsNum int
}

func (st *shopType) GetType() int {
	return st.stype
}

func (st *shopType) GetDirectGoods() (int, int) {
	return st.goodsId, st.goodsNum
}
