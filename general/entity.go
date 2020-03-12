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

func newEntity(run Runtime, entityObject interface{}) *entity {
	e := new(entity)
	e.runtime = run
	eValue := reflect.ValueOf(e)
	ok := false
	allFields(entityObject, func(value reflect.Value) {
		if value.Kind() == reflect.Interface && eValue.Type().Implements(value.Type()) {
			value.Set(eValue)
			ok = true
		}
		return
	})
	if !ok {
		panic("Entities must be inherited from 'freedom.Entity'")
	}
	return e
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
