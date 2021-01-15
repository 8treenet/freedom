//Package po generated by 'freedom new-po'
package po

import (
	"github.com/jinzhu/gorm"
	"time"
)

// Order .
type Order struct {
	changes    map[string]interface{}
	ID         int       `gorm:"primary_key;column:id"`
	OrderNo    string    `gorm:"column:order_no"`
	UserID     int       `gorm:"column:user_id"`     // 用户id
	TotalPrice int       `gorm:"column:total_price"` // 总价
	Status     string    `gorm:"column:status"`      // 支付,未支付，发货，完成
	Created    time.Time `gorm:"column:created"`
	Updated    time.Time `gorm:"column:updated"`
}

// TableName .
func (obj *Order) TableName() string {
	return "order"
}

// Location .
func (obj *Order) Location() map[string]interface{} {
	return map[string]interface{}{"id": obj.ID}
}

// GetChanges .
func (obj *Order) GetChanges() map[string]interface{} {
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

// update .
func (obj *Order) update(name string, value interface{}) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.changes[name] = value
}

// SetOrderNo .
func (obj *Order) SetOrderNo(orderNo string) {
	obj.OrderNo = orderNo
	obj.update("order_no", orderNo)
}

// SetUserID .
func (obj *Order) SetUserID(userID int) {
	obj.UserID = userID
	obj.update("user_id", userID)
}

// SetTotalPrice .
func (obj *Order) SetTotalPrice(totalPrice int) {
	obj.TotalPrice = totalPrice
	obj.update("total_price", totalPrice)
}

// SetStatus .
func (obj *Order) SetStatus(status string) {
	obj.Status = status
	obj.update("status", status)
}

// SetCreated .
func (obj *Order) SetCreated(created time.Time) {
	obj.Created = created
	obj.update("created", created)
}

// SetUpdated .
func (obj *Order) SetUpdated(updated time.Time) {
	obj.Updated = updated
	obj.update("updated", updated)
}

// AddUserID .
func (obj *Order) AddUserID(userID int) {
	obj.UserID += userID
	obj.update("user_id", gorm.Expr("user_id + ?", userID))
}

// AddTotalPrice .
func (obj *Order) AddTotalPrice(totalPrice int) {
	obj.TotalPrice += totalPrice
	obj.update("total_price", gorm.Expr("total_price + ?", totalPrice))
}
