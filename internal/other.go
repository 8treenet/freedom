package internal

import "reflect"

func newOther() *other {
	result := new(other)
	result.pool = make(map[reflect.Type]reflect.Value)
	return result
}

type other struct {
	installs []func() interface{}
	pool     map[reflect.Type]reflect.Value
}

func (o *other) add(f func() interface{}) {
	o.installs = append(o.installs, f)
}

func (o *other) booting() {
	for i := 0; i < len(o.installs); i++ {
		install := o.installs[i]()
		t := reflect.ValueOf(install).Type()
		for t.Kind() != reflect.Struct {
			t = t.Elem()
		}

		o.pool[t] = reflect.ValueOf(install)
	}
}

func (o *other) get(object interface{}) {
	value := reflect.ValueOf(object)
	vtype := value.Type()
	for vtype.Kind() != reflect.Struct {
		vtype = vtype.Elem()
	}

	poolValue, ok := o.pool[vtype]
	if !ok {
		panic("Not found other")
	}
	value.Elem().Set(poolValue)
}
