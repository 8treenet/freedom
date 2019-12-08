package general

import (
	"bytes"
	"fmt"
	"reflect"
)

func newMessageBus() *MessageBus {
	return &MessageBus{
		msgsPath: make(map[string]string),
	}
}

// MessageBus .
type MessageBus struct {
	msgsPath map[string]string
}

// add .
func (msgBus *MessageBus) add(topic string, controller interface{}, funName string) {
	err := fmt.Sprintf("ListenMessage '%s' panic, the method format must be 'PostMsg{}By(messageKey string)'", funName)
	objType := reflect.TypeOf(controller)

	bytefunName := []byte(funName)
	if len(bytefunName) < 10 || !bytes.Equal([]byte("PostMsg"), bytefunName[0:7]) || !bytes.Equal([]byte("By"), bytefunName[len(bytefunName)-2:]) {
		panic(err)
	}
	for index := 0; index < objType.NumMethod(); index++ {
		if funName != objType.Method(index).Name {
			continue
		}
		ft := objType.Method(index).Func.Type()
		if ft.NumIn() != 2 {
			panic(err)
		}
		if ft.In(1).Kind() != reflect.String {
			panic(err)
		}
		msgBus.msgsPath[topic] = objType.Elem().String() + "." + funName
		return
	}
	panic("ListenMessage panic, The method was not found :" + funName)
}

// MessagesPath .
func (msgBus *MessageBus) MessagesPath() (msgs map[string]string) {
	msgs = make(map[string]string)
	for _, r := range globalApp.IrisApp.GetRoutes() {
		if r.Method != "POST" {
			continue
		}

		for topic, handler := range msgBus.msgsPath {
			if r.MainHandlerName == handler {
				msgs[topic] = r.Path
			}
		}
	}
	return
}
