package domainevent

import (
	"database/sql"

	dt "DT-Go"
	"DT-Go/infra/transaction"
)

func init() {
	dt.Prepare(func(initiator dt.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *EventTransaction {
			return &EventTransaction{}
		})
	})
}

const workerStorePubEventKey = "WORKER_STORE_PUB_EVENT_KEY"

// EventTransaction .
type EventTransaction struct {
	transaction.SqlDBImpl
}

// Execute .
func (et *EventTransaction) Execute(f func() error) (err error) {
	defer func() {
		et.Worker().Store().Remove(workerStorePubEventKey)
	}()

	if err = et.SqlDBImpl.Execute(f); err != nil {
		return
	}
	et.pushEvent()
	return
}

// ExecuteTX .
func (et *EventTransaction) ExecuteTX(f func() error, opts *sql.TxOptions) (err error) {
	defer func() {
		et.Worker().Store().Remove(workerStorePubEventKey)
	}()

	if err = et.SqlDBImpl.ExecuteTx(f, opts); err != nil {
		return
	}
	et.pushEvent()
	return
}

// pushEvent 发布事件 使用manager推送
func (et *EventTransaction) pushEvent() {
	pubs := et.Worker().Store().Get(workerStorePubEventKey)
	if pubs == nil {
		return
	}

	pubEvents, ok := pubs.([]dt.DomainEvent)
	if !ok {
		return
	}

	for _, pubEvent := range pubEvents {
		eventManager.push(pubEvent)
	}
	return
}
