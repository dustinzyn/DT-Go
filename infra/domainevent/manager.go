/**
domainevent 领域事件组件

Created by Dustin.zhu on 2022/11/1.
*/

package domainevent

import (
	"Hive"
	"errors"
	"time"

	"gorm.io/gorm"
)

var eventManager *EventManager

func init() {
	eventManager = &EventManager{pubChan: make(chan hive.DomainEvent, 1000)}
	hive.Prepare(func(initiator hive.Initiator) {
		// 单例
		initiator.BindInfra(true, initiator.IsPrivate(), eventManager)
		// InjectController
		initiator.InjectController(func(ctx hive.Context) (com *EventManager) {
			initiator.GetInfra(ctx, &com)
			return
		})
	})
}

// GetEventManager .
func GetEventManager() *EventManager {
	return eventManager
}

type EventManager struct {
	hive.Infra
	pubChan chan hive.DomainEvent
}

// GetPubChan .
func (m *EventManager) GetPubChan() <-chan hive.DomainEvent {
	return m.pubChan
}

// Booting .
func (m *EventManager) Booting(singleBoot hive.SingleBoot) {
	db := m.db()
	if !db.Migrator().HasTable(&domainEventPublish{}) {
		db.AutoMigrate(&domainEventPublish{})
	}

	if !db.Migrator().HasTable(&domainEventSubscribe{}) {
		db.AutoMigrate(&domainEventSubscribe{})
	}
}

// Save .
func (m *EventManager) Save(repo *hive.Repository, entity hive.Entity) (err error) {
	txDB := getTxDB(repo)

	// 删除实体里的全部事件
	defer entity.RemoveAllPubEvent()
	defer entity.RemoveAllSubEvent()

	// Insert PubEvent
	for _, domainEvent := range entity.GetPubEvent() {
		model := domainEventPublish{
			Topic:   domainEvent.Topic(),
			Content: string(domainEvent.Marshal()),
			Created: time.Now(),
			Updated: time.Now(),
		}
		err = txDB.Create(&model).Error
		if err != nil {
			m.Worker().Logger().Errorf("Insert PubEvent error: %v", err)
			return
		}
		domainEvent.SetIdentity(model.ID)
	}
	m.addPubToWOrker(repo.Worker(), entity.GetPubEvent())

	// Update SubEvent
	for _, subEvent := range entity.GetSubEvent() {
		eventID := subEvent.Identity().(int)
		subscribe := &domainEventSubscribe{PublishID: eventID}
		subscribe.SetSuccess(1)
		rowResult := txDB.Model(subscribe).Updates(subscribe.TakeChanges())

		err = rowResult.Error
		if err != nil {
			hive.Logger().Error(err)
			return
		}
		if rowResult.RowsAffected == 0 {
			err = errors.New("Event not found")
			return
		}
	}
	return
}

// InsertSubEvent .
func (m *EventManager) InsertSubEvent(event hive.DomainEvent) error {
	model := domainEventSubscribe{
		PublishID: event.Identity().(int),
		Topic:     event.Topic(),
		Content:   string(event.Marshal()),
		Created:   time.Now(),
		Updated:   time.Now(),
	}
	err := m.db().Create(&model).Error
	if err != nil {
		m.Worker().Logger().Errorf("InsertSubEvent error: %v", err)
		return err
	}
	return nil
}

// Retry 定时器扫描表中失败的Pub/Sub事件
func (m *EventManager) Retry() {
	hive.Logger().Info("EventManager Retry")
	// TODO
}

// push EventTransaction事务成功后触发
func (m *EventManager) push(event hive.DomainEvent) {
	hive.Logger().Infof("Domain Event Topic: %v, %v", event.Topic(), event)
	eventID := event.Identity().(int)
	go func() {
		/**
		发布消息可采用 MQ、Http、RPC、Go Chanel等，这里采用Go Channel
		*/
		m.pubChan <- event
		// push 成功后删除事件
		publish := &domainEventPublish{ID: eventID}
		if err := m.db().Delete(&publish).Error; err != nil {
			hive.Logger().Error(err)
		}
	}()
}

// addPubToWorker 增加发布事件到worker的store
func (m *EventManager) addPubToWOrker(worker hive.Worker, pubs []hive.DomainEvent) {
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

func (m *EventManager) db() *gorm.DB {
	return m.SourceDB().(*gorm.DB)
}

func getTxDB(repo *hive.Repository) (db *gorm.DB) {
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return
}
