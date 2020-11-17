package aggregate

import (
	"encoding/json"

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

// MarshalJSON .
func (query *CartItemQuery) MarshalJSON() ([]byte, error) {
	var obj struct {
		TotalPrice int //总价
		Items      []struct {
			ID         int    //购物车项ID
			GoodsID    int    //商品id
			GoodsName  string //商品名称
			GoodsNum   int    //商品数量
			TotalPrice int    //商品价格
		}
	}
	obj.TotalPrice = query.AllItemTotalPrice()
	for i := 0; i < len(query.allCart); i++ {
		goodsEntity := query.goodsMap[query.allCart[i].GoodsID]
		obj.Items = append(obj.Items, struct {
			ID         int
			GoodsID    int
			GoodsName  string
			GoodsNum   int
			TotalPrice int
		}{
			query.allCart[i].ID,
			goodsEntity.ID,
			goodsEntity.Name,
			query.allCart[i].Num,
			query.allCart[i].Num * goodsEntity.Price,
		})
	}

	return json.Marshal(obj)
}
