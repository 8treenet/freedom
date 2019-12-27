package general

import (
	"reflect"
	"unsafe"
)

func newMessageBus() *EventBus {
	return &EventBus{
		eventsPath:     make(map[string][]string),
		eventsAddr:     make(map[string][]int),
		eventsInfraCom: make(map[string]reflect.Type),
	}
}

// EventBus .
type EventBus struct {
	eventsPath     map[string][]string
	eventsAddr     map[string][]int
	controllers    []interface{}
	eventsInfraCom map[string]reflect.Type
}

func (msgBus *EventBus) addEvent(fun interface{}, eventName string, infraCom ...interface{}) {
	point := (*int)(unsafe.Pointer(&fun))
	msgBus.eventsAddr[eventName] = append(msgBus.eventsAddr[eventName], *point)
	if len(infraCom) > 0 {
		infraComType := reflect.TypeOf(infraCom[0])
		msgBus.eventsInfraCom[eventName] = infraComType
	}
}
func (msgBus *EventBus) addController(controller interface{}) {
	msgBus.controllers = append(msgBus.controllers, controller)
}

// EventsPath .
func (msgBus *EventBus) EventsPath(infra interface{}) (msgs map[string][]string) {
	infraComType := reflect.TypeOf(infra)
	msgs = make(map[string][]string)
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
	eventsRoute := make(map[string][]string)

	for _, controller := range msgBus.controllers {
		v := reflect.ValueOf(controller)
		t := reflect.TypeOf(controller)
		for index := 0; index < v.NumMethod(); index++ {
			funInterface := v.Method(index).Interface()
			point := (*int)(unsafe.Pointer(&funInterface))
			eventName := msgBus.match(*point)
			if eventName == "" {
				continue
			}
			method := t.Method(index)
			ft := method.Func.Type()
			if ft.NumIn() != 2 {
				panic("Event routing parameters must be 'eventID string', MainHandlerName:" + t.Elem().String() + "." + method.Name)
			}
			if ft.In(1).Kind() != reflect.String {
				panic("Event routing parameters must be 'eventID string', MainHandlerName:" + t.Elem().String() + "." + method.Name)
			}
			eventsRoute[eventName] = append(eventsRoute[eventName], t.Elem().String()+"."+method.Name)
		}
	}
	for _, r := range globalApp.IrisApp.GetRoutes() {
		for eventName, handlersName := range eventsRoute {
			for index := 0; index < len(handlersName); index++ {
				if r.MainHandlerName != handlersName[index] {
					continue
				}
				if r.Method != "POST" {
					panic("Event routing must be 'post', MainHandlerName:" + r.MainHandlerName)
				}
				msgBus.eventsPath[eventName] = append(msgBus.eventsPath[eventName], r.Path)
			}
		}
	}
}

func (msgBus *EventBus) match(method int) (eventName string) {
	for name, addrs := range msgBus.eventsAddr {
		for _, addr := range addrs {
			if addr == method {
				eventName = name
				return
			}
		}
	}
	return
}
