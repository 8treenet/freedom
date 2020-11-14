package internal

import (
	"fmt"
	"reflect"
	"strings"

	uuid "github.com/iris-contrib/go.uuid"
)

var _ Entity = (*entity)(nil)

// DomainEvent .
type DomainEvent interface {
	Topic() string
	SetPrototypes(map[string]interface{})
	GetPrototypes() map[string]interface{}
	Marshal() []byte
	Identity() interface{}
	SetIdentity(identity interface{})
}

//Entity is the entity's father interface.
type Entity interface {
	Identity() string
	GetWorker() Worker
	Marshal() []byte
	AddPubEvent(DomainEvent)
	GetPubEvent() []DomainEvent
	RemoveAllPubEvent()
	AddSubEvent(DomainEvent)
	GetSubEvent() []DomainEvent
	RemoveAllSubEvent()
}

func injectBaseEntity(run Worker, entityObject interface{}) {
	entityObjectValue := reflect.ValueOf(entityObject)
	if entityObjectValue.Kind() == reflect.Ptr {
		entityObjectValue = entityObjectValue.Elem()
	}
	entityField := entityObjectValue.FieldByName("Entity")
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
		panic(fmt.Sprintf("[Freedom] InjectBaseEntity: This is not a legitimate entity, %v", entityObjectValue.Type()))
	}
	entityField.Set(eValue)
	return
}

type entity struct {
	worker       Worker
	entityName   string
	identity     string
	producer     string
	entityObject interface{}
	pubEvents    []DomainEvent
	subEvents    []DomainEvent
}

/*
func (e *entity) DomainEvent(fun string, object interface{}, header ...map[string]string) {
	if globalApp.eventInfra == nil {
		globalApp.Logger().Fatalf("[Freedom] DomainEvent: Unrealized Domain Event Infrastructure, %v", reflect.TypeOf(object))
	}
	json, err := globalApp.marshal(object)
	if err != nil {
		panic(err)
	}
	globalApp.eventInfra.DomainEvent(e.producer, fun, json, e.worker, header...)
}
*/

func (e *entity) Identity() string {
	if e.identity == "" {
		u, _ := uuid.NewV1()
		e.identity = strings.ToLower(strings.ReplaceAll(u.String(), "-", ""))
	}
	return e.identity
}

// GetWorker .
func (e *entity) GetWorker() Worker {
	return e.worker
}

// Marshal .
func (e *entity) Marshal() []byte {
	data, err := globalApp.marshal(e.entityObject)
	if err != nil {
		e.worker.Logger().Errorf("[Freedom] Entity.Marshal: serialization failed, %v, error:%v", reflect.TypeOf(e.entityObject), err)
	}
	return data
}

// AddPubEvent.
func (e *entity) AddPubEvent(event DomainEvent) {
	e.pubEvents = append(e.pubEvents, event)
}

// GetPubEvent .
func (e *entity) GetPubEvent() (result []DomainEvent) {
	return e.pubEvents
}

// RemoveAllPubEvent .
func (e *entity) RemoveAllPubEvent() {
	e.pubEvents = []DomainEvent{}
}

// AddSubEvent.
func (e *entity) AddSubEvent(event DomainEvent) {
	e.subEvents = append(e.subEvents, event)
}

// GetSubEvent .
func (e *entity) GetSubEvent() (result []DomainEvent) {
	return e.subEvents
}

// RemoveAllSubEvent .
func (e *entity) RemoveAllSubEvent() {
	e.subEvents = []DomainEvent{}
}
