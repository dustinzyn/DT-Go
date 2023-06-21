/**
domainevent 领域事件组件

Created by Dustin.zhu on 2022/11/1.
*/

package domainevent

import (
	"errors"
	"time"

	hive "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/uniqueid"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/utils"
	"gorm.io/gorm"
)

//go:generate mockgen -package mock_infra -source manager.go -destination ./mock/domainevent_mock.go

var (
	eventManager *EventManagerImpl
)

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
	// Save 保存领域发布事件
	Save(repo *hive.Repository, entity hive.Entity) (err error)
	// InsertSubEvent 插入领域订阅事件
	InsertSubEvent(event hive.DomainEvent) error
	// SetSubEventFail 将订阅事件置为失败状态
	SetSubEventFail(event hive.DomainEvent) error
	// RetryPubThread 定时器扫描表中失败的Pub事件
	RetryPubThread(app hive.Application)
}

type EventManagerImpl struct {
	hive.Infra
	uniqueID   uniqueid.Sonyflaker                      // 唯一性ID组件
	pubHandler func(topic string, content string) error // 发布事件函数 由使用方自定义
}

// Booting .
func (m *EventManagerImpl) Booting(singleBoot hive.SingleBoot) {
	db := m.db()
	if !db.Migrator().HasTable(&domainEventPublish{}) {
		db.AutoMigrate(&domainEventPublish{})
	}

	if !db.Migrator().HasTable(&domainEventSubscribe{}) {
		db.AutoMigrate(&domainEventSubscribe{})
	}
}

// RegisterPubHandler .
func (m *EventManagerImpl) RegisterPubHandler(f func(topic string, content string) error) {
	m.pubHandler = f
}

// Save .
func (m *EventManagerImpl) Save(repo *hive.Repository, entity hive.Entity) (err error) {
	txDB := getTxDB(repo)

	// 删除实体里的全部事件
	defer entity.RemoveAllPubEvent()
	defer entity.RemoveAllSubEvent()

	// Insert PubEvent
	for _, domainEvent := range entity.GetPubEvent() {
		uid, _ := m.uniqueID.NextID()
		model := domainEventPublish{
			ID:      int(uid),
			Topic:   domainEvent.Topic(),
			Content: string(domainEvent.Marshal()),
			Created: time.Now(),
			Updated: time.Now(),
		}
		err = txDB.Create(&model).Error
		if err != nil {
			hive.Logger().Errorf("Insert PubEvent error: %v", err)
			return
		}
		domainEvent.SetIdentity(model.ID)
	}
	m.addPubToWOrker(repo.Worker(), entity.GetPubEvent())

	// Delete SubEvent
	for _, subEvent := range entity.GetSubEvent() {
		eventID := subEvent.Identity().(int)
		subscribe := &domainEventSubscribe{ID: eventID}
		err = txDB.Delete(subscribe).Error

		if err != nil {
			hive.Logger().Error(err)
			return
		}
	}
	return
}

// InsertSubEvent .
func (m *EventManagerImpl) InsertSubEvent(event hive.DomainEvent) error {
	model := domainEventSubscribe{
		ID:      event.Identity().(int),
		Topic:   event.Topic(),
		Content: string(event.Marshal()),
		Created: time.Now(),
		Updated: time.Now(),
	}
	err := m.db().Create(&model).Error
	if err != nil {
		hive.Logger().Errorf("InsertSubEvent error: %v", err)
		return err
	}
	return nil
}

// SetSubEventFail 将订阅事件置为失败状态
func (m *EventManagerImpl) SetSubEventFail(event hive.DomainEvent) error {
	sub := domainEventSubscribe{ID: event.Identity().(int)}
	sub.SetStatus(1)
	sub.SetUpdated(time.Now())
	changes := sub.TakeChanges()
	err := m.db().Model(&sub).Updates(changes).Error
	if err != nil {
		hive.Logger().Errorf("SetSubEventFail error: %v", err)
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
				publish.SetUpdated(time.Now())
				changes := publish.TakeChanges()
				if changes != nil {
					if e := m.db().Model(publish).Updates(changes).Error; e != nil {
						hive.Logger().Errorf("update event error:%v", e)
					}
				}
				return
			}
			// push 成功后删除事件
			if err := m.db().Delete(&publish).Error; err != nil {
				hive.Logger().Error(err)
				return
			}
		}()
		publish = &domainEventPublish{ID: eventID}
		// 发布事件
		err = m.pubHandler(event.Topic(), string(event.Marshal()))
	}()
}

func (m *EventManagerImpl) db() *gorm.DB {
	return m.SourceDB().(*gorm.DB)
}

func getTxDB(repo *hive.Repository) (db *gorm.DB) {
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return
}

// RetryPubThread 定时器扫描表中失败的Pub事件
func (m *EventManagerImpl) RetryPubThread(app hive.Application) {
	time.Sleep(5 * time.Second) //延迟，等待程序Application.Run
	hive.Logger().Info("***************** EventManager Retry Publish *****************")
	timeTicker := time.NewTicker(time.Duration(300) * time.Second)
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

	pubs := make([]domainEventPublish, 0)
	err = m.db().Model(&domainEventPublish{}).Where("status = ?", 1).Scan(&pubs).Limit(100).Error
	if len(pubs) == 0 {
		// 全部完成由定时器再次触发
		needTimer = true
		return
	}
	if err != nil {
		hive.Logger().Errorf("retry pub error:%v", err)
		return
	}
	for _, event := range pubs {
		err = m.pubHandler(event.Topic, event.Content)
		if err != nil {
			hive.Logger().Errorf("execPush error: %v", err)
			continue
		}
		// 推送成功删除事件
		m.db().Delete(&domainEventPublish{ID: event.ID})
	}
	return
}
