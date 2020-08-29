package api_test

import (
	"testing"

	"github.com/8treenet/freedom/example/fshop/domain/dto"
	"github.com/8treenet/freedom/infra/requests"
)

// 获取分页订单
func TestOrderItems(t *testing.T) {
	str, rep := requests.NewH2CRequest("http://127.0.0.1:8000/order/items").Get().SetParam("page", 1).SetParam("pageSize", 5).SetParam("userId", 1).ToString()
	t.Log(str, rep)
}

// 支付订单
func TestOrderPay(t *testing.T) {
	req := dto.OrderPayReq{UserID: 1, OrderNo: "1596885638"}
	str, rep := requests.NewH2CRequest("http://127.0.0.1:8000/order/pay").Put().SetJSONBody(req).ToString()
	t.Log(str, rep)
}
