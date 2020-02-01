package general

import (
	"encoding/json"
	"reflect"
	"strings"
	"unsafe"

	uuid "github.com/iris-contrib/go.uuid"
)

var _ Entity = new(entity)

type Entity interface {
	DomainEvent(fun interface{}, object interface{}, header ...map[string]string)
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
	if globalApp.eventInfra != nil {
		e.funcRouting(entityObject)
	}
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
	funTopics  map[int]string
}

func (e *entity) DomainEvent(fun interface{}, object interface{}, header ...map[string]string) {
	if globalApp.eventInfra == nil {
		panic("Unrealized Domain Event Infrastructure.")
	}

	topic := e.getTopic(fun)
	json, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	globalApp.eventInfra.DomainEvent(e.producer, topic, json, e.runtime, header...)
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

func (e *entity) funcRouting(object interface{}) {
	e.funTopics = make(map[int]string)
	objectValue := reflect.ValueOf(object)
	objectType := reflect.TypeOf(object)
	for index := 0; index < objectValue.NumMethod(); index++ {
		funInterface := objectValue.Method(index).Interface()
		point := (*int)(unsafe.Pointer(&funInterface))
		method := objectType.Method(index)
		e.funTopics[*point] = method.Name
	}

	e.entityName = objectType.Elem().Name()
	return
}

func (e *entity) getTopic(fun interface{}) string {
	point := (*int)(unsafe.Pointer(&fun))
	topic, ok := e.funTopics[*point]
	if !ok {
		panic("Unknown error")
	}
	return e.entityName + ":" + topic
}

func (e *entity) SetProducer(producer string) {
	e.producer = producer
}
