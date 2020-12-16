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
		globalApp.Logger().Fatalf("[Freedom] BindRepository: func return to empty, %v", reflect.TypeOf(fun))
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

func (pool *RepositoryPool) diRepo(repo interface{}, instance *serviceInstance) {
	allFields(repo, func(value reflect.Value) {
		pool.diRepoFromValue(value, instance)
	})
}

func (pool *RepositoryPool) diRepoFromValue(value reflect.Value, instance *serviceInstance) {
	//如果是指针的成员变量
	if value.Kind() == reflect.Ptr && value.IsZero() {
		ok, newfield := pool.get(value.Type())
		if !ok {
			return
		}
		if !value.CanSet() {
			globalApp.IrisApp.Logger().Fatalf("[Freedom] This use repository object must be a capital variable: %v" + value.Type().String())
		}
		//创建实例并且注入基础设施组件
		value.Set(newfield)
		allFieldsFromValue(newfield, func(repoValue reflect.Value) {
			globalApp.comPool.diInfraFromValue(repoValue)
			if br, ok := repoValue.Interface().(BeginRequest); ok {
				instance.calls = append(instance.calls, br)
			}
		})
		//globalApp.comPool.diInfra(newfield.Interface())
	}

	//如果是接口的成员变量
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
				globalApp.IrisApp.Logger().Fatalf("[Freedom] This use repository object must be a capital variable: %v" + value.Type().String())
			}
			//创建实例并且注入基础设施组件
			value.Set(newfield)
			allFieldsFromValue(newfield, func(repoValue reflect.Value) {
				globalApp.comPool.diInfraFromValue(repoValue)
				if br, ok := repoValue.Interface().(BeginRequest); ok {
					instance.calls = append(instance.calls, br)
				}
			})
			//globalApp.comPool.diInfra(newfield.Interface())
			return
		}
	}
}
