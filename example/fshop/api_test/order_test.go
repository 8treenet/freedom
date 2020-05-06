package api_test

import (
	"testing"

	"github.com/8treenet/freedom/example/fshop/application/dto"
	"github.com/8treenet/freedom/general/requests"
)

// 获取分页订单
func TestOrderItems(t *testing.T) {
	str, rep := requests.NewH2CRequest("http://127.0.0.1:8000/order/items").Get().SetParam("page", 1).SetParam("pageSize", 5).SetParam("userId", 1).ToString()
	t.Log(str, rep)
}

// 支付订单
func TestOrderPay(t *testing.T) {
	req := dto.OrderPayReq{UserId: 1, OrderNo: "1588770191"}
	str, rep := requests.NewH2CRequest("http://127.0.0.1:8000/order/pay").Put().SetJSONBody(req).ToString()
	t.Log(str, rep)
}
