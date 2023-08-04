package domainevent

import (
	"fmt"
	"strings"
)

// domainEventSubscribe .
type domainEventSubscribe struct {
	changes map[string]interface{}
	ID      int
	Topic   string // 主题
	Status  int    // 0未处理，1处理失败
	Content string // 内容
	Created int64
	Updated int64
}

// TableName .
func (obj *domainEventSubscribe) TableName() string {
	return "domain_event_subscribe"
}

// TakeChanges .
func (obj *domainEventSubscribe) TakeChanges() (result string) {
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
func (obj *domainEventSubscribe) SetCreated(created int64) {
	obj.Created = created
	obj.setChanges("created", created)
}

// SetUpdated .
func (obj *domainEventSubscribe) SetUpdated(updated int64) {
	obj.Updated = updated
	obj.setChanges("updated", updated)
}
