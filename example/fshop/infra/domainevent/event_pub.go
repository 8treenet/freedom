package domainevent

import (
	"time"

	"gorm.io/gorm"
)

// pubEventObject .
type pubEventObject struct {
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
func (obj *pubEventObject) TableName() string {
	return "pub_event_" + time.Now().Format("200601")
}

// Location .
func (obj *pubEventObject) Location() map[string]interface{} {
	return map[string]interface{}{"identity": obj.Identity}
}

// GetChanges .
func (obj *pubEventObject) GetChanges() map[string]interface{} {
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
func (obj *pubEventObject) update(name string, value interface{}) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.changes[name] = value
}

// AddRetries .
func (obj *pubEventObject) AddRetries(retries int) {
	obj.Retries += retries
	obj.update("retries", gorm.Expr("retries + ?", retries))
}
