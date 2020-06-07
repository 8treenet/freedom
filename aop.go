package freedom

import (
	"fmt"
	"reflect"
)

type proxyInterface interface {
	Call(fun string, args ...interface{}) ([]interface{}, error)
	setTarget(target interface{})
}

type ProxyHandle struct {
	target interface{}
}

func (aopHandle *ProxyHandle) Call(fun string, args ...interface{}) (result []interface{}, e error) {
	value := reflect.ValueOf(aopHandle.target)
	for value.Kind() != reflect.Ptr {
		value = value.Elem()
	}

	call := value.MethodByName(fun)
	if !call.IsValid() {
		return nil, fmt.Errorf("Method %s not found", fun)
	}

	var in []reflect.Value
	for i := 0; i < len(args); i++ {
		in = append(in, reflect.ValueOf(args[i]))
	}
	out := call.Call(in)
	for i := 0; i < len(out); i++ {
		result = append(result, out[i].Interface())
	}

	return result, nil
}

func (aopHandle *ProxyHandle) setTarget(target interface{}) {
	aopHandle.target = target
}

func NewProxy(proxyObj proxyInterface, target interface{}) {
	proxyObj.setTarget(target)
}
