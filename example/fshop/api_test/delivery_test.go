package api_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/8treenet/freedom/example/fshop/domain/dto"
	"github.com/8treenet/freedom/infra/requests"
)

// 发货
func TestDelivery(t *testing.T) {
	var req dto.DeliveryReq
	req.AdminId = 1
	req.TrackingNumber = fmt.Sprint(rand.Intn(999999999999999))
	req.OrderNo = "1596885638"

	str, rep := requests.NewH2CRequest("http://127.0.0.1:8000/delivery").Post().SetJSONBody(req).ToString()
	t.Log(str, rep)
}
