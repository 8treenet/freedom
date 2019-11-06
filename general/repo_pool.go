package general

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
		panic("BindService func return to empty")
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
