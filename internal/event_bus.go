package internal

import (
	"reflect"
)

func newMessageBus() *EventBus {
	return &EventBus{
		eventsPath:     make(map[string]string),
		eventsAddr:     make(map[string]string),
		eventsInfraCom: make(map[string]reflect.Type),
	}
}

// EventBus .
type EventBus struct {
	eventsPath     map[string]string
	eventsAddr     map[string]string
	controllers    []interface{}
	eventsInfraCom map[string]reflect.Type
}

func (msgBus *EventBus) addEvent(objectMethod, eventName string, infraCom ...interface{}) {
	if _, ok := msgBus.eventsAddr[eventName]; ok {
		panic("Event already bound :" + eventName)
	}
	msgBus.eventsAddr[eventName] = objectMethod
	if len(infraCom) > 0 {
		infraComType := reflect.TypeOf(infraCom[0])
		msgBus.eventsInfraCom[eventName] = infraComType
	}
}
func (msgBus *EventBus) addController(controller interface{}) {
	msgBus.controllers = append(msgBus.controllers, controller)
}

// EventsPath .
func (msgBus *EventBus) EventsPath(infra interface{}) (msgs map[string]string) {
	infraComType := reflect.TypeOf(infra)
	msgs = make(map[string]string)
	for k, v := range msgBus.eventsPath {
		ty, ok := msgBus.eventsInfraCom[k]
		if ok && ty != infraComType {
			continue
		}
		msgs[k] = v
	}
	return
}

func (msgBus *EventBus) building() {
	eventsRoute := make(map[string]string)
	for _, controller := range msgBus.controllers {
		v := reflect.ValueOf(controller)
		t := reflect.TypeOf(controller)
		for index := 0; index < v.NumMethod(); index++ {
			method := t.Method(index)
			eventName := msgBus.match(t.Elem().Name() + "." + method.Name)
			if eventName == "" {
				continue
			}
			ft := method.Func.Type()
			if ft.NumIn() != 2 {
				panic("Event routing parameters must be 'eventID string', MainHandlerName:" + t.Elem().String() + "." + method.Name)
			}
			if ft.In(1).Kind() != reflect.String {
				panic("Event routing parameters must be 'eventID string', MainHandlerName:" + t.Elem().String() + "." + method.Name)
			}
			eventsRoute[eventName] = t.Elem().String() + "." + method.Name
		}
	}
	for _, r := range globalApp.IrisApp.GetRoutes() {
		for eventName, handlersName := range eventsRoute {
			if r.MainHandlerName != handlersName {
				continue
			}
			if r.Method != "POST" {
				panic("Event routing must be 'post', MainHandlerName:" + r.MainHandlerName)
			}
			msgBus.eventsPath[eventName] = r.Path
		}
	}
}

func (msgBus *EventBus) match(objectMethod string) (eventName string) {
	for name, addrs := range msgBus.eventsAddr {
		if addrs == objectMethod {
			eventName = name
			return
		}
	}
	return
}
