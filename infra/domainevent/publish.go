package domainevent

import (
	"fmt"
	"strings"
)

// domainEventPublish .
type domainEventPublish struct {
	changes map[string]interface{}
	ID      int
	Topic   string // 主题
	Content string // 事件内容
	Status  int    // 0:待处理 1:处理失败
	Created int64
	Updated int64
}

// TableName .
func (obj *domainEventPublish) TableName() string {
	return "domain_event_publish"
}

// TakeChanges .
func (obj *domainEventPublish) TakeChanges() (result string) {
	if obj.changes == nil {
		return ""
	}
	for k, v := range obj.changes {
		result += fmt.Sprintf("%v=%v,", k, v)
	}
	result = strings.TrimRight(result, ",")
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
func (obj *domainEventPublish) SetCreated(created int64) {
	obj.Created = created
	obj.setChanges("created", created)
}

// SetStatus .
func (obj *domainEventPublish) SetStatus(status int) {
	obj.Status = status
	obj.setChanges("status", status)
}

// SetUpdated .
func (obj *domainEventPublish) SetUpdated(updated int64) {
	obj.Updated = updated
	obj.setChanges("updated", updated)
}
