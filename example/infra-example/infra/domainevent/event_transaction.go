package domainevent

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/infra/transaction"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindInfra(false, func() *EventTransaction {
			return &EventTransaction{} //基于请求的多例绑定
		})
	})
}

const (
	workerStorePubEventKey = "WORKER_STORE_PUB_EVENT_KEY"
)

//EventTransaction .
type EventTransaction struct {
	transaction.GormImpl
}

// Execute .
func (et *EventTransaction) Execute(fun func() error) (e error) {
	defer func() {
		et.Worker().Store().Remove(workerStorePubEventKey)
	}()

	if e = et.GormImpl.Execute(fun); e != nil {
		return
	}
	et.pushEvent()
	return
}

// pushEvent 发布事件 .
func (et *EventTransaction) pushEvent() {
	pubs := et.Worker().Store().Get(workerStorePubEventKey)
	if pubs == nil {
		return
	}
	pubEvents, ok := pubs.([]freedom.DomainEvent)
	if !ok {
		return
	}

	for _, pubEvent := range pubEvents {
		eventManager.push(pubEvent) //使用manager推送
	}
}
