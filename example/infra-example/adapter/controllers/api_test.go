package controllers

import (
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

// 购买商品
func TestPostShop(t *testing.T) {
	var request struct {
		GoodsID int `json:"goodsId"` //商品id
		Num     int `json:"num"`     //购买数量
		UserID  int `json:"userId"`  //用户id
	}
	request.GoodsID = 1
	request.Num = 15
	request.UserID = 1001

	str, resp := requests.NewHTTPRequest("http://127.0.0.1:8000/shop").Post().SetJSONBody(request).ToString()
	t.Log(str, resp)
}

//查看指定商品
func TestGetGoods(t *testing.T) {
	str, resp := requests.NewHTTPRequest("http://127.0.0.1:8000/goods/2").Get().ToString()
	t.Log(str, resp)
}

//查看全部商品
func TestGetGoodsList(t *testing.T) {
	str, resp := requests.NewHTTPRequest("http://127.0.0.1:8000/goods").Get().ToString()
	t.Log(str, resp)
}

//增加库存
func TestPutStock(t *testing.T) {
	var request struct {
		GoodsID int `json:"goodsId"` //商品id
		Num     int `json:"num"`     //增加数量
	}
	request.GoodsID = 1
	request.Num = 14
	//request.Num = 5

	str, resp := requests.NewHTTPRequest("http://127.0.0.1:8000/goods/stock").Put().SetJSONBody(request).ToString()
	t.Log(str, resp)
}

//查看指定订单
func TestGetOrder(t *testing.T) {
	str, resp := requests.NewHTTPRequest("http://127.0.0.1:8000/order/1").Get().SetParam("userId", 1001).ToString()
	t.Log(str, resp)
}

//查看全部订单
func TestGetOrders(t *testing.T) {
	str, resp := requests.NewHTTPRequest("http://127.0.0.1:8000/order").Get().SetParam("userId", 1001).ToString()
	t.Log(str, resp)
}
