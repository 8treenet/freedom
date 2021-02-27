package internal

import (
	"fmt"
	"reflect"
)

func newCustom() *custom {
	result := new(custom)
	result.pool = make(map[reflect.Type]reflect.Value)
	return result
}

type custom struct {
	installs []func() interface{}
	pool     map[reflect.Type]reflect.Value
}

func (c *custom) add(f func() interface{}) {
	c.installs = append(c.installs, f)
}

func (c *custom) booting() {
	for i := 0; i < len(c.installs); i++ {
		install := c.installs[i]()
		t := reflect.ValueOf(install).Type()
		for t.Kind() != reflect.Struct {
			t = t.Elem()
		}

		c.pool[t] = reflect.ValueOf(install)
	}
}

func (c *custom) get(object interface{}) {
	value := reflect.ValueOf(object)
	vtype := value.Type()
	for vtype.Kind() != reflect.Struct {
		vtype = vtype.Elem()
	}

	poolValue, ok := c.pool[vtype]
	if !ok {
		panic(fmt.Sprintf("[Freedom] Repository.Other: Does not exist, %v", vtype))
	}
	value.Elem().Set(poolValue)
}
