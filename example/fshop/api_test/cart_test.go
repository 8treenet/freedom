package api_test

import (
	"testing"

	"github.com/8treenet/freedom/example/fshop/domain/dto"
	"github.com/8treenet/freedom/infra/requests"
)

// 获取全部购物车商品
func TestCartGetItems(t *testing.T) {
	str, err := requests.NewH2CRequest("http://127.0.0.1:8000/cart/items").SetQueryParam("userId", 1).Get().ToString()
	t.Log(str, err)
}

// 添加购物车商品
func TestPostCart(t *testing.T) {
	var req dto.CartAddReq
	req.UserID = 1
	req.GoodsID = 3
	req.GoodsNum = 1
	str, err := requests.NewH2CRequest("http://127.0.0.1:8000/cart").Post().SetJSONBody(req).ToString()
	t.Log(str, err)

	req.GoodsID = 2
	req.GoodsNum = 1
	str2, err2 := requests.NewH2CRequest("http://127.0.0.1:8000/cart").Post().SetJSONBody(req).ToString()
	t.Log(str2, err2)
}

// 清空购物车
func TestDeleteCarts(t *testing.T) {
	str, err := requests.NewH2CRequest("http://127.0.0.1:8000/cart/all").SetQueryParam("userId", 1).Delete().ToString()
	t.Log(str, err)
}

// 购买购物车商品
func TestCartShop(t *testing.T) {
	obj := dto.CartShopReq{
		UserID: 1, //用户id
	}
	str, err := requests.NewHTTPRequest("http://127.0.0.1:8000/cart/shop").Post().SetJSONBody(obj).ToString()
	t.Log(str, err)
}
