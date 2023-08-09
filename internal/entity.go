package internal

import (
	"fmt"
	"reflect"
)

var _ Entity = (*entity)(nil)

// DomainEvent.
type DomainEvent interface {
	Topic() string
	SetPrototypes(map[string]interface{})
	GetPrototypes() map[string]interface{}
	Marshal() []byte
	Identity() interface{}
	SetIdentity(identity interface{})
}

// Entity is the entity's father interface.
type Entity interface {
	Identity() int
	Worker() Worker
	Marshal() []byte
	AddPubEvent(DomainEvent)
	GetPubEvent() []DomainEvent
	RemoveAllPubEvent()
	AddSubEvent(DomainEvent)
	GetSubEvent() []DomainEvent
	RemoveAllSubEvent()
}

type entity struct {
	worker       Worker
	entityName   string
	identity     int
	producer     string
	entityObject interface{}
	pubEvents    []DomainEvent
	subEvents    []DomainEvent
}

// injectBaseEntity
func injectBaseEntity(run Worker, entityObject interface{}) {
	entityObjValue := reflect.ValueOf(entityObject)
	if entityObjValue.Kind() == reflect.Ptr {
		entityObjValue = entityObjValue.Elem()
	}
	entityField := entityObjValue.FieldByName("Entity")
	if !entityField.IsNil() {
		return
	}

	e := new(entity)
	e.worker = run
	e.entityObject = entityObject
	e.pubEvents = []DomainEvent{}
	e.subEvents = []DomainEvent{}
	eValue := reflect.ValueOf(e)
	if entityField.Kind() != reflect.Interface || !eValue.Type().Implements(entityField.Type()) {
		panic(fmt.Sprintf("InjectBaseEntity: This is not a legitimate entity, %v", entityObjValue.Type()))
	}
	entityField.Set(eValue)
	return
}

func (e *entity) Identity() int {
	return e.identity
}

func (e *entity) Worker() Worker {
	return e.worker
}

func (e *entity) Marshal() []byte {
	data, err := e.app().marshal(e.entityObject)
	if err != nil {
		e.worker.Logger().Errorf("Entity.Marshal: serialization failed, %v, error: %v", reflect.TypeOf(e.entityObject), err)
	}
	return data
}

func (e *entity) AddPubEvent(event DomainEvent) {
	e.pubEvents = append(e.pubEvents, event)
}

func (e *entity) GetPubEvent() (result []DomainEvent) {
	return e.pubEvents
}

func (e *entity) RemoveAllPubEvent() {
	e.pubEvents = []DomainEvent{}
}

func (e *entity) AddSubEvent(event DomainEvent) {
	e.subEvents = append(e.subEvents, event)
}

func (e *entity) GetSubEvent() (result []DomainEvent) {
	return e.subEvents
}

func (e *entity) RemoveAllSubEvent() {
	e.subEvents = []DomainEvent{}
}

// app returns an application
func (e *entity) app() *Application {
	if e.worker.IsPrivate() {
		return privateApp
	} else {
		return publicApp
	}
}
