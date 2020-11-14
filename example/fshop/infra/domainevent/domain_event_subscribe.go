package domainevent

import (
	"time"
)

// domainEventSubscribe .
type domainEventSubscribe struct {
	changes   map[string]interface{}
	ID        int       `gorm:"primary_key;column:id;auto increment"`
	PublishID int       `gorm:"unique;not null;column:publish_id"`
	Topic     string    `gorm:"column:topic;not null"`             // 主题
	Success   int       `gorm:"column:success;not null"`           // 0未处理，1处理
	Content   string    `gorm:"column:content;size:2000;not null"` // 内容
	Created   time.Time `gorm:"column:created;not null"`
	Updated   time.Time `gorm:"column:updated;not null"`
}

// TableName .
func (obj *domainEventSubscribe) TableName() string {
	return "domain_event_subscribe"
}

// TakeChanges .
func (obj *domainEventSubscribe) TakeChanges() map[string]interface{} {
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
func (obj *domainEventSubscribe) setChanges(name string, value interface{}) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.changes[name] = value
}

// SetTopic .
func (obj *domainEventSubscribe) SetTopic(topic string) {
	obj.Topic = topic
	obj.setChanges("topic", topic)
}

// SetSuccess .
func (obj *domainEventSubscribe) SetSuccess(success int) {
	obj.Success = success
	obj.setChanges("success", success)
}

// SetContent .
func (obj *domainEventSubscribe) SetContent(content string) {
	obj.Content = content
	obj.setChanges("content", content)
}

// SetCreated .
func (obj *domainEventSubscribe) SetCreated(created time.Time) {
	obj.Created = created
	obj.setChanges("created", created)
}

// SetUpdated .
func (obj *domainEventSubscribe) SetUpdated(updated time.Time) {
	obj.Updated = updated
	obj.setChanges("updated", updated)
}
