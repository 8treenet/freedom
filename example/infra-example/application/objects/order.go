package objects

import (
	"time"
)

// Order .
type Order struct {
	changes map[string]interface{}
	ID      int       `gorm:"primary_key" column:"id"`
	UserID  int       `gorm:"column:user_id"`  // 用户id
	GoodsID int       `gorm:"column:goods_id"` // 商品id
	Num     int       `gorm:"column:num"`      // 数量
	Created time.Time `gorm:"column:created"`
	Updated time.Time `gorm:"column:updated"`
}

func (obj *Order) TableName() string {
	return "order"
}

// Save .
func (obj *Order) Save() map[string]interface{} {
	if obj.changes == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range obj.changes {
		result[k] = v
	}
	obj.changes = nil
	return result
}

// SetID .
func (obj *Order) SetID(iD int) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.ID = iD
	obj.changes["id"] = iD
}

// SetUserID .
func (obj *Order) SetUserID(userID int) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.UserID = userID
	obj.changes["user_id"] = userID
}

// SetGoodsID .
func (obj *Order) SetGoodsID(goodsID int) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.GoodsID = goodsID
	obj.changes["goods_id"] = goodsID
}

// SetNum .
func (obj *Order) SetNum(num int) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.Num = num
	obj.changes["num"] = num
}

// SetCreated .
func (obj *Order) SetCreated(created time.Time) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.Created = created
	obj.changes["created"] = created
}

// SetUpdated .
func (obj *Order) SetUpdated(updated time.Time) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.Updated = updated
	obj.changes["updated"] = updated
}
