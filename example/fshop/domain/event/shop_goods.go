package event

import "encoding/json"

// ShopGoods 购买事件
type ShopGoods struct {
	ID         int `json:"id"`
	prototypes map[string]interface{}
	UserID     int    `json:"userID"`
	OrderNO    string `json:"orderNO"`
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
func (shop *ShopGoods) Marshal() []byte {
	data, _ := json.Marshal(shop)
	return data
}

// Identity .
func (shop *ShopGoods) Identity() interface{} {
	return shop.ID
}

// SetIdentity .
func (shop *ShopGoods) SetIdentity(identity interface{}) {
	shop.ID = identity.(int)
}
