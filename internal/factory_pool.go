package internal

import "reflect"

func newFactoryPool() *FactoryPool {
	result := new(FactoryPool)
	result.creater = make(map[reflect.Type]interface{})
	return result
}

// FactoryPool .
type FactoryPool struct {
	creater map[reflect.Type]interface{}
}

func (pool *FactoryPool) bind(outType reflect.Type, f interface{}) {
	pool.creater[outType] = f
}

// get .
func (pool *FactoryPool) get(t reflect.Type) (ok bool, result reflect.Value) {
	fun, ok := pool.creater[t]
	if !ok {
		return false, reflect.ValueOf(nil)
	}

	values := reflect.ValueOf(fun).Call([]reflect.Value{})
	if len(values) == 0 {
		globalApp.IrisApp.Logger().Fatalf("[freedom]BindFactory func return to empty, %v", t)
	}

	return true, values[0]
}

func (pool *FactoryPool) exist(t reflect.Type) bool {
	_, ok := pool.creater[t]
	return ok
}

// allType .
func (pool *FactoryPool) allType() (list []reflect.Type) {
	for t := range pool.creater {
		list = append(list, t)
	}
	return
}

// diFactory .
func (pool *FactoryPool) diFactory(entity interface{}) {
	allFields(entity, func(value reflect.Value) {
		//如果是指针的成员变量
		if value.Kind() == reflect.Ptr && value.IsZero() {
			ok, newfield := pool.get(value.Type())
			if !ok {
				return
			}
			if !value.CanSet() {
				globalApp.IrisApp.Logger().Fatalf("[freedom]This use factory object must be a capital variable: %v" + value.Type().String())
			}
			//创建实例并且注入基础设施组件和资源库
			value.Set(newfield)
			factoryObj := newfield.Interface()
			globalApp.comPool.diInfra(factoryObj)
			globalApp.rpool.diRepo(factoryObj)
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
					globalApp.IrisApp.Logger().Fatalf("[freedom]This use factory object must be a capital variable: %v", value.Type().String())
				}
				//创建实例并且注入基础设施组件和资源库
				value.Set(newfield)
				factoryObj := newfield.Interface()
				globalApp.comPool.diInfra(factoryObj)
				globalApp.rpool.diRepo(factoryObj)
				return
			}
		}
	})
}
