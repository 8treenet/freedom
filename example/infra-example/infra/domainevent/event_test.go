package domainevent

import (
	"encoding/json"
	"testing"
)

// shopGoods 购买事件
type shopGoods struct {
	ID         string `json:"identity"`
	prototypes map[string]interface{}
	UserID     int `json:"userID"`
	GoodsID    int `json:"goodsID"`
}

// Topic .
func (shop *shopGoods) Topic() string {
	return "ShopGoods"
}

// Marshal .
func (shop *shopGoods) Marshal() ([]byte, error) {
	return json.Marshal(shop)
}

// Unmarshal .
func (shop *shopGoods) Unmarshal(data []byte) error {
	return json.Unmarshal(data, shop)
}

// SetPrototypes .
func (shop *shopGoods) SetPrototypes(prototypes map[string]interface{}) {
}

// GetPrototypes .
func (shop *shopGoods) GetPrototypes() map[string]interface{} {
	return shop.prototypes
}

// Identity .
func (shop *shopGoods) Identity() string {
	return shop.ID
}

// SetIdentity .
func (shop *shopGoods) SetIdentity(identity string) {
	shop.ID = identity
}

func TestRetry(t *testing.T) {
	GetEventManager().BindRetrySubEvent(&shopGoods{}, func(event *shopGoods) {
		t.Log("fuck", event)
	})

	data, _ := (&shopGoods{
		ID:      "111222",
		UserID:  1,
		GoodsID: 234,
	}).Marshal()

	po := subEventObject{
		changes: map[string]interface{}{
			"": nil,
		},
		Sequence: 0,
		Identity: "111222",
		Topic:    "ShopGoods",
		Content:  string(data),
	}
	GetEventManager().eventRetry.callSub(&po)
}
