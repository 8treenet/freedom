package internal

import (
	"fmt"
	"reflect"

	uuid "github.com/iris-contrib/go.uuid"
)

var _ Entity = (*entity)(nil)

// DomainEvent Interface definition of domain events for subscription and publishing of entities.
// To create a domain event, you must implement the interface's methods.
type DomainEvent interface {
	//The only noun used to identify an event.
	Topic() string
	//Set the property.
	SetPrototypes(map[string]interface{})
	//Read the property.
	GetPrototypes() map[string]interface{}
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	//The unique ID of the event.
	Identity() string
	//Set a unique ID.
	SetIdentity(identity string)
}

//Entity is the entity's father interface.
type Entity interface {
	//The unique ID of the entity.
	Identity() string
	//Returns the requested Worer
	Worker() Worker
	Marshal() ([]byte, error)
	//Add a publishing event.
	AddPubEvent(DomainEvent)
	//Get all publishing events..
	GetPubEvents() []DomainEvent
	//Delete all publishing events.
	RemoveAllPubEvent()
	//Add a subscription event.
	AddSubEvent(DomainEvent)
	//Get all subscription events..
	GetSubEvents() []DomainEvent
	//Delete all subscription events.
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
		e.identity = u.String()
	}
	return e.identity
}

// Worker .
func (e *entity) Worker() Worker {
	return e.worker
}

// Marshal .
func (e *entity) Marshal() ([]byte, error) {
	return globalApp.marshal(e.entityObject)
}

// AddPubEvent.
func (e *entity) AddPubEvent(event DomainEvent) {
	if reflect.ValueOf(event.Identity()).IsZero() {
		//如果未设置id
		u, err := uuid.NewV1()
		if err != nil {
			panic(err)
		}
		event.SetIdentity(u.String())
	}

	e.pubEvents = append(e.pubEvents, event)
}

// GetPubEvent .
func (e *entity) GetPubEvents() (result []DomainEvent) {
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
func (e *entity) GetSubEvents() (result []DomainEvent) {
	return e.subEvents
}

// RemoveAllSubEvent .
func (e *entity) RemoveAllSubEvent() {
	e.subEvents = []DomainEvent{}
}
