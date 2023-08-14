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

	hive "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/config"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/uniqueid"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/utils"
	"devops.aishu.cn/AISHUDevOps/ONE-Architecture/_git/proton-rds-sdk-go/sqlx"
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
	hive.Prepare(func(initiator hive.Initiator) {
		// 单例
		initiator.BindInfra(true, initiator.IsPrivate(), eventManager)
		// InjectController
		initiator.InjectController(func(ctx hive.Context) (com *EventManagerImpl) {
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
	RetryPubEvent(app hive.Application)
	// RetrySubEvent 定时器扫描表中失败的Pub事件
	RetrySubEvent(app hive.Application)
	// Save 保存领域事件
	Save(repo *hive.Repository, entity hive.Entity) (err error)
	// DeleteSubEvent 删除领域订阅事件
	DeleteSubEvent(eventID int) error
	// SetSubEventFail 将订阅事件置为失败状态
	SetSubEventFail(eventID int) error
}

type EventManagerImpl struct {
	hive.Infra
	uniqueID   uniqueid.Sonyflaker                      // 唯一性ID组件
	pubHandler func(topic string, content string) error // 发布事件函数 由使用方自定义
	subHandler func(topic string, content string) error // 订阅事件函数 由使用方自定义
}

// Booting .
func (m *EventManagerImpl) Booting(singleBoot hive.SingleBoot) {
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
func (m *EventManagerImpl) RetryPubEvent(app hive.Application) {
	time.Sleep(time.Duration(DelayInterval) * time.Second) //延迟，等待程序Application.Run
	hive.Logger().Info("***************** EventManager Retry Publish *****************")
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
func (m *EventManagerImpl) RetrySubEvent(app hive.Application) {
	time.Sleep(time.Duration(DelayInterval) * time.Second) // 延迟，等待程序Application.Run
	hive.Logger().Info("***************** EventManager Retry Subscribe *****************")
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
func (m *EventManagerImpl) Save(repo *hive.Repository, entity hive.Entity) (err error) {
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
		_, err = txDB.Exec(sqlStr, uid, domainEvent.Topic(), string(domainEvent.Marshal()), ct, ct, 0)
		if err != nil {
			hive.Logger().Errorf("Insert PubEvent error: %v", err)
			return
		}
		domainEvent.SetIdentity(uid)
	}
	m.addPubToWOrker(repo.Worker(), entity.GetPubEvent())

	// Insert SubEvent
	for _, subEvent := range entity.GetSubEvent() {
		sqlStr := "INSERT INTO %v.domain_event_subscribe (id, topic, content, created, updated, status) VALUES (?, ?, ?, ?, ?, ?)"
		sqlStr = fmt.Sprintf(sqlStr, m.dbConfig().DBName)
		_, err := m.db().Exec(sqlStr, subEvent.Identity().(int), subEvent.Topic(), string(subEvent.Marshal()), ct, ct, 0)
		if err != nil {
			hive.Logger().Errorf("InsertSubEvent error: %v", err)
			return err
		}
	}
	return
}

// DeleteSubEvent 删除领域订阅事件
func (m *EventManagerImpl) DeleteSubEvent(eventID int) error {
	sqlStr := "DELETE FROM %v.domain_event_subscribe WHERE id = ?"
	sqlStr = fmt.Sprintf(sqlStr, m.dbConfig().DBName)
	_, err := m.db().Exec(sqlStr, eventID)
	if err != nil {
		hive.Logger().Errorf("DeleteSubEvent error: %v", err)
		return err
	}
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
		_, err = m.db().Exec(sqlStr, changes, eventID)
		if err != nil {
			hive.Logger().Errorf("SetSubEventFail error: %v", err)
			return err
		}
	}
	return nil
}

func (m *EventManagerImpl) retrySub() (needTimer bool) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			hive.Logger().Errorf("retrySub error: %v", r)
			err = errors.New("retrySub panic, recover")
		}
		if err != nil {
			needTimer = true
			return
		}
	}()
	subs, err := eventManager.getFailSubEvents(SingleRetryNum)
	if err != nil {
		hive.Logger().Infof("retrySub error: %v", err)
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
			hive.Logger().Errorf("execPush error: %v", err)
			continue
		}
		// 推送成功删除事件
		if err = eventManager.DeleteSubEvent(event["id"].(int)); err != nil {
			hive.Logger().Errorf("execPush: delete sub event error: %v", err)
		}
	}
	return
}

func (m *EventManagerImpl) retryPub() (needTimer bool) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			hive.Logger().Errorf("retryPub error: %v", r)
			err = errors.New("retryPub panic, recover")
		}
		if err != nil {
			needTimer = true
			return
		}
	}()

	pubs, err := m.getFailPubEvents(SingleRetryNum)
	if err != nil {
		hive.Logger().Infof("retryPub error: %v", err)
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
			hive.Logger().Errorf("execPush error: %v", err)
			continue
		}
		// 推送成功删除事件
		if err = eventManager.DeletePubEvent(event["id"].(int)); err != nil {
			hive.Logger().Errorf("execPush: delete pub event error: %v", err)
		}
	}
	return
}

// getFailSubEvents 获取n个处理失败的领域订阅事件(id,topic,content)
func (m *EventManagerImpl) getFailSubEvents(n int) (subs []map[string]interface{}, err error) {
	subs = make([]map[string]interface{}, 0)
	sqlStr := "SELECT id, topic, content FROM %v.domain_event_subscribe WHERE status = ? LIMIT ?"
	sqlStr = fmt.Sprintf(sqlStr, m.dbConfig().DBName)
	rows, err := m.db().Query(sqlStr, 1, n)
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
	sqlStr := "SELECT id, topic, content FROM %v.domain_event_publish WHERE status = ? LIMIT ?"
	sqlStr = fmt.Sprintf(sqlStr, m.dbConfig().DBName)
	rows, err := m.db().Query(sqlStr, 1, n)
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
	_, err := m.db().Exec(sqlStr, eventID)
	if err != nil {
		hive.Logger().Errorf("DeletePubEvent error: %v", err)
		return err
	}
	return nil
}

// addPubToWorker 增加发布事件到worker的store
func (m *EventManagerImpl) addPubToWOrker(worker hive.Worker, pubs []hive.DomainEvent) {
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
	var storePubEvents []hive.DomainEvent
	store := worker.Store().Get(workerStorePubEventKey)
	if store != nil {
		if list, ok := store.([]hive.DomainEvent); ok {
			storePubEvents = list
		}
	}
	storePubEvents = append(storePubEvents, pubs...)
	worker.Store().Set(workerStorePubEventKey, storePubEvents)
}

// push EventTransaction事务成功后触发
func (m *EventManagerImpl) push(event hive.DomainEvent) {
	hive.Logger().Infof("PubEventID: %v, Domain PubEvent Topic: %v, Content: %v", event.Identity(), event.Topic(), event)
	eventID := event.Identity().(int)
	go func() {
		var err error
		var publish *domainEventPublish
		defer func() {
			if r := recover(); r != nil {
				err = errors.New("push panic, recover")
				hive.Logger().Errorf("event push error: %v", r)
			}
			if err != nil {
				// 推送失败 标记事件为失败
				hive.Logger().Errorf("push event error:%v", err)
				publish.SetStatus(1)
				publish.SetUpdated(utils.NowTimestamp())
				changes := publish.TakeChanges()
				if changes != "" {
					sqlStr := "UPDATE %v.domain_event_publish SET ? WHERE id = ?"
					sqlStr = fmt.Sprintf(sqlStr, m.dbConfig().DBName)
					if _, e := m.db().Exec(sqlStr, changes, eventID); e != nil {
						hive.Logger().Errorf("update event error:%v", e)
					}
				}
				return
			}
			// push 成功后删除事件
			sqlStr := "DELETE FROM %v.domain_event_publish WHERE id = ?"
			sqlStr = fmt.Sprintf(sqlStr, m.dbConfig().DBName)
			if _, err := m.db().Exec(sqlStr, eventID); err != nil {
				hive.Logger().Error(err)
				return
			}
		}()
		publish = &domainEventPublish{ID: eventID}
		// 发布事件
		err = m.pubHandler(event.Topic(), string(event.Marshal()))
	}()
}

func (m *EventManagerImpl) closeRows(rows *sql.Rows) {
	if rows != nil {
		if rowsErr := rows.Err(); rowsErr != nil {
			hive.Logger().Error(rowsErr)
		}

		if closeErr := rows.Close(); closeErr != nil {
			hive.Logger().Error(closeErr)
		}
	}
}

func (m *EventManagerImpl) dbConfig() *hive.DBConfiguration {
	return config.NewConfiguration().DB
}

func (m *EventManagerImpl) db() *sqlx.DB {
	return m.SourceDB().(*sqlx.DB)
}

func getTxDB(repo *hive.Repository) (db *sql.Tx) {
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return
}
