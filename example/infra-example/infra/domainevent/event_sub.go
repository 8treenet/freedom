package domainevent

import (
	"time"

	"gorm.io/gorm"
)

// subEventObject .
type subEventObject struct {
	changes  map[string]interface{}
	Sequence int       `gorm:"primaryKey;column:sequence;autoIncrement"`
	Identity string    `gorm:"unique;not null;column:identity"`
	Topic    string    `gorm:"column:topic;not null"`             // 主题
	Retries  int       `gorm:"column:retries;not null"`           // 重试次数
	Content  string    `gorm:"column:content;size:4000;not null"` // 内容
	Created  time.Time `gorm:"column:created;type:TIMESTAMP;default:CURRENT_TIMESTAMP;not null"`
	Updated  time.Time `gorm:"column:updated;type:TIMESTAMP;default:CURRENT_TIMESTAMP;not null"`
}

// TableName .
func (obj *subEventObject) TableName() string {
	return "sub_event_" + time.Now().Format("200601")
}

// Location .
func (obj *subEventObject) Location() map[string]interface{} {
	return map[string]interface{}{"identity": obj.Identity}
}

// GetChanges .
func (obj *subEventObject) GetChanges() map[string]interface{} {
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

// updateChanges .
func (obj *subEventObject) update(name string, value interface{}) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.changes[name] = value
}

// AddRetries .
func (obj *subEventObject) AddRetries(retries int) {
	obj.Retries += retries
	obj.update("retries", gorm.Expr("retries + ?", retries))
}
