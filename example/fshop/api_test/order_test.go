package api_test

import (
	"testing"

	"github.com/8treenet/freedom/example/fshop/domain/vo"
	"github.com/8treenet/freedom/infra/requests"
)

// 获取分页订单
func TestOrderItems(t *testing.T) {
	str, rep := requests.NewH2CRequest("http://127.0.0.1:8000/order/items").Get().SetQueryParam("page", 1).SetQueryParam("pageSize", 5).SetQueryParam("userId", 1).ToString()
	t.Log(str, rep)
}

// 支付订单
func TestOrderPay(t *testing.T) {
	req := vo.OrderPayReq{UserID: 1, OrderNo: "1599896625"}
	str, rep := requests.NewH2CRequest("http://127.0.0.1:8000/order/pay").Put().SetJSONBody(req).ToString()
	t.Log(str, rep)
}
