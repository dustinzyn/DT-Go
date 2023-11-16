/**
domainevent 领域事件组件

Created by Dustin.zhu on 2022/11/1.
*/

package domainevent

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	dt "DT-Go"
	"DT-Go/config"
	"DT-Go/infra/uniqueid"
	"DT-Go/utils"

	"gorm.io/gorm"
)

//go:generate mockgen -package mock_infra -source manager.go -destination ./mock/domainevent_mock.go

const (
	// DelayInterval 延迟启动间隔
	DelayInterval int = 5
	// RetryInterval 重试扫描失败的事件间隔
	RetryInterval int = 30
	// SingleRetryNum 每次重试读取的sub/pub事件数
	SingleRetryNum int = 100
)

var (
	eventManager *EventManagerImpl
)
var _ EventManager = (*EventManagerImpl)(nil)

func init() {
	eventManager = &EventManagerImpl{}
	uniqueID := &uniqueid.SonyflakerImpl{}
	uniqueID.SetPodIP(utils.GetEnv("POD_IP", "127.0.0.1"))
	eventManager.uniqueID = uniqueID
	dt.Prepare(func(initiator dt.Initiator) {
		// 单例
		initiator.BindInfra(true, initiator.IsPrivate(), eventManager)
		// InjectController
		initiator.InjectController(func(ctx dt.Context) (com *EventManagerImpl) {
			initiator.GetInfra(ctx, &com)
			return
		})
		// 绑定资源库
		initiator.BindRepository(func() *EventManagerImpl {
			return &EventManagerImpl{}
		})
	})
}

// GetEventManager .
func GetEventManager() *EventManagerImpl {
	return eventManager
}

type EventManager interface {
	// RegisterPubHandler 注册领域发布事件函数
	RegisterPubHandler(f func(topic string, content string) error)
	// RegisterSubHandler 注册领域发布事件函数
	RegisterSubHandler(f func(topic string, content string) error)
	// RetryPubEvent 定时器扫描表中失败的Pub事件
	RetryPubEvent(app dt.Application)
	// RetrySubEvent 定时器扫描表中失败的Pub事件
	RetrySubEvent(app dt.Application)
	// Save 保存领域事件
	Save(repo *dt.Repository, entity dt.Entity) (err error)
	// DeleteSubEvent 删除领域订阅事件
	DeleteSubEvent(eventID int) error
	// SetSubEventFail 将订阅事件置为失败状态
	SetSubEventFail(eventID int) error
}

type EventManagerImpl struct {
	dt.Infra
	uniqueID   uniqueid.Sonyflaker                      // 唯一性ID组件
	pubHandler func(topic string, content string) error // 发布事件函数 由使用方自定义
	subHandler func(topic string, content string) error // 订阅事件函数 由使用方自定义
}

// Booting .
func (m *EventManagerImpl) Booting(singleBoot dt.SingleBoot) {
	return
}

// RegisterPubHandler .
func (m *EventManagerImpl) RegisterPubHandler(f func(topic string, content string) error) {
	m.pubHandler = f
}

// RegisterSubHandler .
func (m *EventManagerImpl) RegisterSubHandler(f func(topic string, content string) error) {
	m.subHandler = f
}

// RetryPubEvent 定时器扫描表中失败的Pub事件
func (m *EventManagerImpl) RetryPubEvent(app dt.Application) {
	time.Sleep(time.Duration(DelayInterval) * time.Second) //延迟，等待程序Application.Run
	dt.Logger().Info("***************** EventManager Retry Publish *****************")
	timeTicker := time.NewTicker(time.Duration(RetryInterval) * time.Second)
	needTimer := true
	for {
		select {
		case <-timeTicker.C:
			if !needTimer {
				continue
			}
		}
		for {
			needTimer = m.retryPub()
			// 全部重试成功后退出 定时器会再次触发
			if needTimer {
				break
			}
		}
	}
}

// RetrySubEvent 定时器扫描表中失败的Sub事件
func (m *EventManagerImpl) RetrySubEvent(app dt.Application) {
	time.Sleep(time.Duration(DelayInterval) * time.Second) // 延迟，等待程序Application.Run
	dt.Logger().Info("***************** EventManager Retry Subscribe *****************")
	timeTicker := time.NewTicker(time.Duration(RetryInterval) * time.Second)
	needTimer := true
	for {
		<-timeTicker.C
		if !needTimer {
			continue
		}
		for {
			needTimer = m.retrySub()
			// 全部重试成功后退出 定时器会再次触发
			if needTimer {
				break
			}
		}
	}
}

// Save 保存领域事件
func (m *EventManagerImpl) Save(repo *dt.Repository, entity dt.Entity) (err error) {
	txDB := getTxDB(repo)

	// 删除实体里的全部事件
	defer entity.RemoveAllPubEvent()
	defer entity.RemoveAllSubEvent()

	ct := utils.NowTimestamp()
	// Insert PubEvent
	for _, domainEvent := range entity.GetPubEvent() {
		uid, _ := m.uniqueID.NextID()
		sqlStr := "INSERT INTO %v.domain_event_publish (id, topic, content, created, updated, status) VALUES (?, ?, ?, ?, ?, ?)"
		sqlStr = fmt.Sprintf(sqlStr, m.dbConfig().DBName)
		txDB.Exec(sqlStr, uid, domainEvent.Topic(), string(domainEvent.Marshal()), ct, ct, 0)
		domainEvent.SetIdentity(uid)
	}
	m.addPubToWOrker(repo.Worker(), entity.GetPubEvent())

	// Insert SubEvent
	for _, subEvent := range entity.GetSubEvent() {
		sqlStr := "INSERT INTO %v.domain_event_subscribe (id, topic, content, created, updated, status) VALUES (?, ?, ?, ?, ?, ?)"
		sqlStr = fmt.Sprintf(sqlStr, m.dbConfig().DBName)
		txDB.Exec(sqlStr, subEvent.Identity(), subEvent.Topic(), string(subEvent.Marshal()), ct, ct, 0)
	}
	return
}

// DeleteSubEvent 删除领域订阅事件
func (m *EventManagerImpl) DeleteSubEvent(eventID int) error {
	sqlStr := "DELETE FROM %v.domain_event_subscribe WHERE id = ?"
	sqlStr = fmt.Sprintf(sqlStr, m.dbConfig().DBName)
	m.db().Exec(sqlStr, eventID)
	return nil
}

// SetSubEventFail 将订阅事件置为失败状态
func (m *EventManagerImpl) SetSubEventFail(eventID int) (err error) {
	sub := domainEventSubscribe{ID: eventID}
	sub.SetStatus(1)
	sub.SetUpdated(utils.NowTimestamp())
	changes := sub.TakeChanges()
	if changes != "" {
		sqlStr := "UPDATE %v.domain_event_subscribe SET ? WHERE id = ?"
		sqlStr = fmt.Sprintf(sqlStr, m.dbConfig().DBName)
		m.db().Exec(sqlStr, changes, eventID)
	}
	return nil
}

func (m *EventManagerImpl) retrySub() (needTimer bool) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			dt.Logger().Errorf("retrySub error: %v", r)
			err = errors.New("retrySub panic, recover")
		}
		if err != nil {
			needTimer = true
			return
		}
	}()
	subs, err := eventManager.getFailSubEvents(SingleRetryNum)
	if err != nil {
		dt.Logger().Infof("retrySub error: %v", err)
		needTimer = true
		return
	}
	if len(subs) == 0 {
		// 全部完成由定时器再次触发
		needTimer = true
		return
	}
	for _, event := range subs {
		err = m.subHandler(event["topic"].(string), event["content"].(string))
		if err != nil {
			dt.Logger().Errorf("execPush error: %v", err)
			continue
		}
		// 推送成功删除事件
		if err = eventManager.DeleteSubEvent(event["id"].(int)); err != nil {
			dt.Logger().Errorf("execPush: delete sub event error: %v", err)
		}
	}
	return
}

func (m *EventManagerImpl) retryPub() (needTimer bool) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			dt.Logger().Errorf("retryPub error: %v", r)
			err = errors.New("retryPub panic, recover")
		}
		if err != nil {
			needTimer = true
			return
		}
	}()

	pubs, err := m.getFailPubEvents(SingleRetryNum)
	if err != nil {
		dt.Logger().Infof("retryPub error: %v", err)
		needTimer = true
		return
	}
	if len(pubs) == 0 {
		// 全部完成由定时器再次触发
		needTimer = true
		return
	}
	for _, event := range pubs {
		err = m.pubHandler(event["topic"].(string), event["content"].(string))
		if err != nil {
			dt.Logger().Errorf("execPush error: %v", err)
			continue
		}
		// 推送成功删除事件
		if err = eventManager.DeletePubEvent(event["id"].(int)); err != nil {
			dt.Logger().Errorf("execPush: delete pub event error: %v", err)
		}
	}
	return
}

// getFailSubEvents 获取n个处理失败的领域订阅事件(id,topic,content)
func (m *EventManagerImpl) getFailSubEvents(n int) (subs []map[string]interface{}, err error) {
	subs = make([]map[string]interface{}, 0)

	rows, err := m.db().Table(fmt.Sprintf("%v.domain_event_subscribe", m.dbConfig().DBName)).Where("status = ?", 1).Limit(n).Rows()
	defer utils.CloseRows(rows)
	if err != nil {
		return
	}
	for rows.Next() {
		var id int
		var topic, content string
		if err = rows.Scan(&id, &topic, &content); err != nil {
			return
		}
		sub := map[string]interface{}{"id": id, "topic": topic, "content": content}
		subs = append(subs, sub)
	}
	return
}

// getFailPubEvents 获取n个处理失败的领域发布事件(id,topic,content)
func (m *EventManagerImpl) getFailPubEvents(n int) (pubs []map[string]interface{}, err error) {
	pubs = make([]map[string]interface{}, 0)
	rows, err := m.db().Table(fmt.Sprintf("%v.domain_event_publish", m.dbConfig().DBName)).Where("status = ?", 1).Limit(n).Rows()
	defer utils.CloseRows(rows)
	if err != nil {
		return
	}
	for rows.Next() {
		var id int
		var topic, content string
		if err = rows.Scan(&id, &topic, &content); err != nil {
			return
		}
		sub := map[string]interface{}{"id": id, "topic": topic, "content": content}
		pubs = append(pubs, sub)
	}
	return
}

// DeletePubEvent 删除领域发布事件
func (m *EventManagerImpl) DeletePubEvent(eventID int) error {
	sqlStr := "DELETE FROM %v.domain_event_publish WHERE id = ?"
	sqlStr = fmt.Sprintf(sqlStr, m.dbConfig().DBName)
	m.db().Exec(sqlStr, eventID)
	return nil
}

// addPubToWorker 增加发布事件到worker的store
func (m *EventManagerImpl) addPubToWOrker(worker dt.Worker, pubs []dt.DomainEvent) {
	if len(pubs) == 0 {
		return
	}

	for _, pubEvent := range pubs {
		m := make(map[string]interface{})
		for key, item := range worker.Bus().Header.Clone() {
			if len(item) <= 0 {
				continue
			}
			m[key] = item[0]
		}
		pubEvent.SetPrototypes(m)
	}

	// 把发布事件添加到store, EventTransaction在事务结束后会触发push
	var storePubEvents []dt.DomainEvent
	store := worker.Store().Get(workerStorePubEventKey)
	if store != nil {
		if list, ok := store.([]dt.DomainEvent); ok {
			storePubEvents = list
		}
	}
	storePubEvents = append(storePubEvents, pubs...)
	worker.Store().Set(workerStorePubEventKey, storePubEvents)
}

// push EventTransaction事务成功后触发
func (m *EventManagerImpl) push(event dt.DomainEvent) {
	dt.Logger().Infof("PubEventID: %v, Domain PubEvent Topic: %v, Content: %v", event.Identity(), event.Topic(), event)
	eventID := event.Identity().(int)
	go func() {
		var err error
		var publish *domainEventPublish
		defer func() {
			if r := recover(); r != nil {
				err = errors.New("push panic, recover")
				dt.Logger().Errorf("event push error: %v", r)
			}
			if err != nil {
				// 推送失败 标记事件为失败
				dt.Logger().Errorf("push event error:%v", err)
				publish.SetStatus(1)
				publish.SetUpdated(utils.NowTimestamp())
				changes := publish.TakeChanges()
				if changes != "" {
					sqlStr := "UPDATE %v.domain_event_publish SET ? WHERE id = ?"
					sqlStr = fmt.Sprintf(sqlStr, m.dbConfig().DBName)
					m.db().Exec(sqlStr, changes, eventID)
				}
				return
			}
			// push 成功后删除事件
			sqlStr := "DELETE FROM %v.domain_event_publish WHERE id = ?"
			sqlStr = fmt.Sprintf(sqlStr, m.dbConfig().DBName)
			m.db().Exec(sqlStr, eventID)
		}()
		publish = &domainEventPublish{ID: eventID}
		// 发布事件
		err = m.pubHandler(event.Topic(), string(event.Marshal()))
	}()
}

func (m *EventManagerImpl) closeRows(rows *sql.Rows) {
	if rows != nil {
		if rowsErr := rows.Err(); rowsErr != nil {
			dt.Logger().Error(rowsErr)
		}

		if closeErr := rows.Close(); closeErr != nil {
			dt.Logger().Error(closeErr)
		}
	}
}

func (m *EventManagerImpl) dbConfig() config.DBConfiguration {
	return *config.NewConfiguration().DB
}

func (m *EventManagerImpl) db() *gorm.DB {
	return m.SourceDB().(*gorm.DB)
}

func getTxDB(repo *dt.Repository) (db *gorm.DB) {
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return
}
