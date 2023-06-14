package domainevent

import (
	"time"
)

// domainEventSubscribe .
type domainEventSubscribe struct {
	changes   map[string]interface{}
	ID        int       `gorm:"primary_key;column:id;auto increment"`
	Topic     string    `gorm:"column:topic;size:50;not null"`     // 主题
	Status    int       `gorm:"column:status;not null"`            // 0未处理，1处理失败
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

// SetStatus .
func (obj *domainEventSubscribe) SetStatus(status int) {
	obj.Status = status
	obj.setChanges("status", status)
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
