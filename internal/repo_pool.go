package internal

import (
	"reflect"
)

func newRepoPool() *RepositoryPool {
	result := new(RepositoryPool)
	result.creater = make(map[reflect.Type]interface{})
	return result
}

// RepositoryPool .
type RepositoryPool struct {
	creater map[reflect.Type]interface{}
}

// get .
func (pool *RepositoryPool) get(t reflect.Type) (ok bool, result reflect.Value) {
	fun, ok := pool.creater[t]
	if !ok {
		return false, reflect.ValueOf(nil)
	}

	values := reflect.ValueOf(fun).Call([]reflect.Value{})
	if len(values) == 0 {
		panic("BindRepository func return to empty")
	}

	return true, values[0]
}

func (pool *RepositoryPool) bind(outType reflect.Type, f interface{}) {
	pool.creater[outType] = f
}

func (pool *RepositoryPool) allType() (list []reflect.Type) {
	for t := range pool.creater {
		list = append(list, t)
	}
	return
}

func (pool *RepositoryPool) diRepo(entity interface{}) {
	allFields(entity, func(value reflect.Value) {
		if value.Kind() == reflect.Ptr && value.IsZero() {
			ok, newfield := pool.get(value.Type())
			if !ok {
				return
			}
			if !value.CanSet() {
				globalApp.IrisApp.Logger().Fatal("The member variable must be publicly visible, Its type is " + value.Type().String())
			}
			value.Set(newfield)
			globalApp.comPool.diInfra(newfield.Interface())
		}

		if value.Kind() == reflect.Interface && value.IsZero() {
			typeList := pool.allType()
			for index := 0; index < len(typeList); index++ {
				if !typeList[index].Implements(value.Type()) {
					continue
				}
				ok, newfield := pool.get(typeList[index])
				if !ok {
					continue
				}
				if !value.CanSet() {
					globalApp.IrisApp.Logger().Fatal("The member variable must be publicly visible, Its type is " + value.Type().String())
				}
				pool.diRepo(newfield.Interface())
				value.Set(newfield)
				globalApp.comPool.diInfra(newfield.Interface())
				return
			}
		}
	})
}
