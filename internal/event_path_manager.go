package internal

import (
	"reflect"
)

func newEventPathManager() *eventPathManager {
	return &eventPathManager{
		eventsPath:     make(map[string]string),
		eventsAddr:     make(map[string]string),
		eventsInfraCom: make(map[string]reflect.Type),
	}
}

// eventPathManager.
// Subscribe to the message conversion HTTP API.
type eventPathManager struct {
	eventsPath     map[string]string
	eventsAddr     map[string]string
	controllers    []interface{}
	eventsInfraCom map[string]reflect.Type
}

func (msgBus *eventPathManager) addEvent(objectMethod, eventName string, infraCom ...interface{}) {
	if _, ok := msgBus.eventsAddr[eventName]; ok {
		globalApp.Logger().Fatalf("[Freedom] ListenEvent: Event already bound :%v", eventName)
	}
	msgBus.eventsAddr[eventName] = objectMethod
	if len(infraCom) > 0 {
		infraComType := reflect.TypeOf(infraCom[0])
		msgBus.eventsInfraCom[eventName] = infraComType
	}
}
func (msgBus *eventPathManager) addController(controller interface{}) {
	msgBus.controllers = append(msgBus.controllers, controller)
}

// EventsPath .
func (msgBus *eventPathManager) EventsPath(infra interface{}) (msgs map[string]string) {
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

func (msgBus *eventPathManager) building() {
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
			eventsRoute[eventName] = t.Elem().String() + "." + method.Name
		}
	}
	for _, r := range globalApp.IrisApp.GetRoutes() {
		for eventName, handlersName := range eventsRoute {
			if r.MainHandlerName != handlersName {
				continue
			}
			if r.Method != "POST" {
				globalApp.Logger().Fatalf("[Freedom] ListenEvent: Event routing must be 'post', MainHandlerName:%v", r.MainHandlerName)
			}
			msgBus.eventsPath[eventName] = r.Path
		}
	}
}

func (msgBus *eventPathManager) match(objectMethod string) (eventName string) {
	for name, addrs := range msgBus.eventsAddr {
		if addrs == objectMethod {
			eventName = name
			return
		}
	}
	return
}
