package internal

import (
	"reflect"
	"strings"

	uuid "github.com/iris-contrib/go.uuid"
)

var _ Entity = (*entity)(nil)

//Entity is the entity's father interface.
type Entity interface {
	DomainEvent(eventName string, object interface{}, header ...map[string]string)
	Identity() string
	GetWorker() Worker
	SetProducer(string)
	Marshal() []byte
}

//DomainEventInfra is a dependency inverted interface for domain events.
type DomainEventInfra interface {
	DomainEvent(producer, topic string, data []byte, worker Worker, header ...map[string]string)
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
	eValue := reflect.ValueOf(e)
	if entityField.Kind() != reflect.Interface || !eValue.Type().Implements(entityField.Type()) {
		globalApp.Logger().Fatalf("[Freedom] InjectBaseEntity: This is not a legitimate entity, %v", entityObjectValue.Type())
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
}

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

func (e *entity) Identity() string {
	if e.identity == "" {
		u, _ := uuid.NewV1()
		e.identity = strings.ToLower(strings.ReplaceAll(u.String(), "-", ""))
	}
	return e.identity
}

func (e *entity) GetWorker() Worker {
	return e.worker
}

func (e *entity) SetProducer(producer string) {
	e.producer = producer
}

func (e *entity) Marshal() []byte {
	data, err := globalApp.marshal(e.entityObject)
	if err != nil {
		e.worker.Logger().Errorf("[Freedom] Entity.Marshal: serialization failed, %v, error:%v", reflect.TypeOf(e.entityObject), err)
	}
	return data
}
