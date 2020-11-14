package api_test

import (
	"testing"

	"github.com/8treenet/freedom/example/fshop/domain/dto"
	"github.com/8treenet/freedom/infra/requests"
)

// 获取分页商品
func TestGetItems(t *testing.T) {
	str, err := requests.NewH2CRequest("http://127.0.0.1:8000/goods/items").Get().SetQueryParam("page", 2).SetQueryParam("pageSize", 2).ToString()
	t.Log(str, err)
	str2, err2 := requests.NewH2CRequest("http://127.0.0.1:8000/goods/items").Get().SetQueryParam("tag", "HOT").SetQueryParam("page", 1).SetQueryParam("pageSize", 2).ToString()
	t.Log(str2, err2)
}

// 创建商品
func TestPostGoods(t *testing.T) {
	var goodsAddReq dto.GoodsAddReq
	goodsAddReq.Name = "freedom"
	goodsAddReq.Price = 50

	str, err := requests.NewH2CRequest("http://127.0.0.1:8000/goods").Post().SetJSONBody(goodsAddReq).ToString()
	t.Log(str, err)
}

// 修改商品库存
func TestGoodsPutStock(t *testing.T) {
	str, err := requests.NewH2CRequest("http://127.0.0.1:8000/goods/stock/2/10").Put().ToString()
	t.Log(str, err)
}

// 商品打标签
func TestGoodsPutTag(t *testing.T) {
	var goodsTagReq struct {
		Id  int
		Tag string //`validate:"oneof=HOT NEW NONE"` //要设置的标签必须是 热门，新品，默认
	}
	goodsTagReq.Id = 5
	goodsTagReq.Tag = "NONE"
	str, err := requests.NewH2CRequest("http://127.0.0.1:8000/goods/tag").Put().SetJSONBody(goodsTagReq).ToString()
	t.Log(str, err)

	goodsTagReq.Tag = "HOT"
	requests.NewH2CRequest("http://127.0.0.1:8000/goods/tag").Put().SetJSONBody(goodsTagReq).ToString()
}

// 购买商品
func TestGoodsShop(t *testing.T) {
	obj := dto.GoodsShopReq{
		UserID: 1, //用户id
		ID:     2, //商品id
		Num:    2, //商品数量
	}
	str, err := requests.NewHTTPRequest("http://127.0.0.1:8000/goods/shop").Post().SetJSONBody(obj).ToString()
	t.Log(str, err)
}
