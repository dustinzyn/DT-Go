package domainevent

import (
	"time"
)

// domainEventPublish .
type domainEventPublish struct {
	changes       map[string]interface{}
	ID            int       `gorm:"primary_key;column:id"`
	Topic         string    `gorm:"column:topic;size:50;not null;comment:'主题'"`
	Content       string    `gorm:"column:content;size:2000;not null;comment:'内容'"`
	Status        int       `gorm:"column:status;not null;comment:'0:待处理 1:处理失败'"`
	Created       time.Time `gorm:"column:created;not null"`
	Updated       time.Time `gorm:"column:updated;not null"`
}

// TableName .
func (obj *domainEventPublish) TableName() string {
	return "domain_event_publish"
}

// TakeChanges .
func (obj *domainEventPublish) TakeChanges() map[string]interface{} {
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
func (obj *domainEventPublish) setChanges(name string, value interface{}) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.changes[name] = value
}

// SetTopic .
func (obj *domainEventPublish) SetTopic(topic string) {
	obj.Topic = topic
	obj.setChanges("topic", topic)
}

// SetContent .
func (obj *domainEventPublish) SetContent(content string) {
	obj.Content = content
	obj.setChanges("content", content)
}

// SetCreated .
func (obj *domainEventPublish) SetCreated(created time.Time) {
	obj.Created = created
	obj.setChanges("created", created)
}

// SetStatus .
func (obj *domainEventPublish) SetStatus(status int) {
	obj.Status = status
	obj.setChanges("status", status)
}

// SetUpdated .
func (obj *domainEventPublish) SetUpdated(updated time.Time) {
	obj.Updated = updated
	obj.setChanges("updated", updated)
}
