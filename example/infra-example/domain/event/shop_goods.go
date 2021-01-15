package event

import (
	"encoding/json"

	"github.com/8treenet/freedom/example/infra-example/infra/domainevent"
)

func init() {
	//不注册 不会触发重试
	domainevent.GetEventManager().RetryPubEvent(&ShopGoods{})
}

// ShopGoods 购买事件
type ShopGoods struct {
	ID         string `json:"identity"`
	prototypes map[string]interface{}
	UserID     int    `json:"userID"`
	GoodsID    int    `json:"goodsID"`
	GoodsNum   int    `json:"goodsNum"`
	GoodsName  string `json:"goodsName"`
}

// Topic .
func (shop *ShopGoods) Topic() string {
	return "ShopGoods"
}

// SetPrototypes .
func (shop *ShopGoods) SetPrototypes(prototypes map[string]interface{}) {
	if shop.prototypes == nil {
		shop.prototypes = prototypes
		return
	}

	for key, value := range prototypes {
		shop.prototypes[key] = value
	}
}

// GetPrototypes .
func (shop *ShopGoods) GetPrototypes() map[string]interface{} {
	return shop.prototypes
}

// Marshal .
func (shop *ShopGoods) Marshal() ([]byte, error) {
	return json.Marshal(shop)
}

// Unmarshal .
func (shop *ShopGoods) Unmarshal(data []byte) error {
	return json.Unmarshal(data, shop)
}

// Identity .
func (shop *ShopGoods) Identity() string {
	return shop.ID
}

// SetIdentity .
func (shop *ShopGoods) SetIdentity(identity string) {
	shop.ID = identity
}
