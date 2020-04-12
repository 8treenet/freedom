package api_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/8treenet/freedom/example/fshop/application/dto"
	"github.com/8treenet/freedom/general/requests"
)

// 发货
func TestDelivery(t *testing.T) {
	var req dto.DeliveryReq
	req.AdminId = 1
	req.TrackingNumber = fmt.Sprint(rand.Intn(999999999999999))
	req.OrderNo = "1586687439"

	str, rep := requests.NewH2CRequest("http://127.0.0.1:8000/delivery").Post().SetJSONBody(req).ToString()
	t.Log(str, rep)
}
