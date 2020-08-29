package aggregate

import (
	"github.com/8treenet/freedom/example/fshop/domain/entity"
)

//CartItemQuery 购物车项查询聚合根
type CartItemQuery struct {
	entity.User
	allCart  []*entity.Cart
	goodsMap map[int]*entity.Goods
}

// VisitAllItem 读取全部商品
func (query *CartItemQuery) VisitAllItem(f func(id, goodsId int, goodsName string, goodsNum, totalPrice int)) {
	for i := 0; i < len(query.allCart); i++ {
		goodsEntity := query.goodsMap[query.allCart[i].GoodsID]
		f(query.allCart[i].ID, goodsEntity.ID, goodsEntity.Name, query.allCart[i].Num, query.allCart[i].Num*goodsEntity.Price)
	}
}

// AllItemTotalPrice 全部商品总价
func (query *CartItemQuery) AllItemTotalPrice() (totalPrice int) {
	for i := 0; i < len(query.allCart); i++ {
		goodsEntity := query.goodsMap[query.allCart[i].GoodsID]
		totalPrice += query.allCart[i].Num * goodsEntity.Price
	}
	return
}
