package general

import (
	"encoding/json"
	"reflect"
	"strings"

	uuid "github.com/iris-contrib/go.uuid"
)

var _ Entity = new(entity)

type Entity interface {
	DomainEvent(eventName string, object interface{}, header ...map[string]string)
	Identity() string
	GetRuntime() Runtime
	SetProducer(string)
}

type DomainEventInfra interface {
	DomainEvent(producer, topic string, data []byte, runtime Runtime, header ...map[string]string)
}

func bindEntity(run Runtime, entityObject interface{}) {
	entityObjectValue := reflect.ValueOf(entityObject)
	if entityObjectValue.Kind() == reflect.Ptr {
		entityObjectValue = entityObjectValue.Elem()
	}
	entityField := entityObjectValue.FieldByName("Entity")
	if !entityField.IsNil() {
		return
	}

	e := new(entity)
	e.runtime = run
	eValue := reflect.ValueOf(e)
	if entityField.Kind() != reflect.Interface || !eValue.Type().Implements(entityField.Type()) {
		panic("This is not a valid entity")
	}
	entityField.Set(eValue)
	return
}

type entity struct {
	runtime    Runtime
	entityName string
	identity   string
	producer   string
}

func (e *entity) DomainEvent(fun string, object interface{}, header ...map[string]string) {
	if globalApp.eventInfra == nil {
		panic("Unrealized Domain Event Infrastructure.")
	}
	json, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	globalApp.eventInfra.DomainEvent(e.producer, fun, json, e.runtime, header...)
}

func (e *entity) Identity() string {
	if e.identity == "" {
		u, _ := uuid.NewV1()
		e.identity = strings.ToLower(strings.ReplaceAll(u.String(), "-", ""))
	}
	return e.identity
}

func (e *entity) GetRuntime() Runtime {
	return e.runtime
}

func (e *entity) SetProducer(producer string) {
	e.producer = producer
}
