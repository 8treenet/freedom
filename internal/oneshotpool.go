package internal

import (
	"fmt"
	"reflect"
)

func newOneShotPool() *oneShotPool {
	return &oneShotPool{
		pool: make(map[reflect.Type]reflect.Value),
	}
}

type oneShotPool struct {
	providers []func() interface{}
	pool      map[reflect.Type]reflect.Value
}

func (o *oneShotPool) addProvider(f func() interface{}) {
	o.providers = append(o.providers, f)
}

func (o *oneShotPool) resolve() {
	for provider := range o.providers {
		t := reflect.ValueOf(provider).Type()
		for t.Kind() != reflect.Struct {
			t = t.Elem()
		}

		o.pool[t] = reflect.ValueOf(provider)
	}
}

func (o *oneShotPool) get(object interface{}) {
	value := reflect.ValueOf(object)
	vtype := value.Type()
	for vtype.Kind() != reflect.Struct {
		vtype = vtype.Elem()
	}

	poolValue, ok := o.pool[vtype]
	if !ok {
		panic(fmt.Sprintf("[Freedom] Repository.Other: Does not exist, %v", vtype))
	}
	value.Elem().Set(poolValue)
}
