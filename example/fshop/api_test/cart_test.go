package api_test

import (
	"testing"

	"github.com/8treenet/freedom/example/fshop/application/dto"
	"github.com/8treenet/freedom/general/requests"
)

// 获取全部购物车商品
func TestCartGetItems(t *testing.T) {
	str, err := requests.NewH2CRequest("http://127.0.0.1:8000/cart/items").SetParam("userId", 1).Get().ToString()
	t.Log(str, err)
}

// 增加购物车商品
func TestPostCart(t *testing.T) {
	var req dto.CartAddReq
	req.UserId = 1
	req.GoodsId = 1
	req.GoodsNum = 1
	str, err := requests.NewH2CRequest("http://127.0.0.1:8000/cart").Post().SetJSONBody(req).ToString()
	t.Log(str, err)

	req.GoodsId = 2
	req.GoodsNum = 1
	str2, err2 := requests.NewH2CRequest("http://127.0.0.1:8000/cart").Post().SetJSONBody(req).ToString()
	t.Log(str2, err2)
}

// 清空购物车
func TestDeleteCarts(t *testing.T) {
	str, err := requests.NewH2CRequest("http://127.0.0.1:8000/cart/all").SetParam("userId", 1).Delete().ToString()
	t.Log(str, err)
}
