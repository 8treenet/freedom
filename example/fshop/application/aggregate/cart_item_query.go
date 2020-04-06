package aggregate

import (
	"github.com/8treenet/freedom/example/fshop/adapter/repository"
	"github.com/8treenet/freedom/example/fshop/application/entity"
)

// CartItemQuery 购物车项查询聚合根，传入相关仓库的接口
func NewCartItemQuery(userRepo repository.UserRepo, cartRepo repository.CartRepo, goodsRepo repository.GoodsRepo) *CartItemQuery {
	return &CartItemQuery{
		userRepo:  userRepo,
		cartRepo:  cartRepo,
		goodsRepo: goodsRepo,
	}
}

// 购物车项查询聚合根
type CartItemQuery struct {
	entity.User
	userRepo  repository.UserRepo
	cartRepo  repository.CartRepo
	goodsRepo repository.GoodsRepo
	allCart   []*entity.Cart
	goodsMap  map[int]*entity.Goods
}

// LoadEntity 加载依赖实体
func (query *CartItemQuery) LoadEntity(userId int) error {
	user, e := query.userRepo.Find(userId)
	if e != nil {
		query.GetRuntime().Logger().Error(e, "userId", userId)
		//用户不存在
		return e
	}

	query.User = *user
	query.goodsMap = make(map[int]*entity.Goods)
	return nil
}

// QueryAllItem 查询购物车内全部商品
func (query *CartItemQuery) QueryAllItem() (e error) {
	query.allCart, e = query.cartRepo.FindAll(query.User.Id)
	for i := 0; i < len(query.allCart); i++ {
		goodsEntity, e := query.goodsRepo.Find(query.allCart[i].GoodsId)
		if e != nil {
			return e
		}
		query.goodsMap[goodsEntity.Id] = goodsEntity
	}
	return
}

// VisitAllItem 读取全部商品
func (query *CartItemQuery) VisitAllItem(f func(id, goodsId int, goodsName string, goodsNum, totalPrice int)) {
	for i := 0; i < len(query.allCart); i++ {
		goodsEntity := query.goodsMap[query.allCart[i].GoodsId]
		f(query.allCart[i].Id, goodsEntity.Id, goodsEntity.Name, query.allCart[i].Num, query.allCart[i].Num*goodsEntity.Price)
	}
}

// AllItemTotalPrice 全部商品总价
func (query *CartItemQuery) AllItemTotalPrice() (totalPrice int) {
	for i := 0; i < len(query.allCart); i++ {
		goodsEntity := query.goodsMap[query.allCart[i].GoodsId]
		totalPrice += query.allCart[i].Num * goodsEntity.Price
	}
	return
}
