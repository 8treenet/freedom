package general

import (
	"reflect"
	"sync"

	"github.com/jinzhu/gorm"
)

// GORMRepository .
type GORMRepository interface {
	DB() *gorm.DB
	InjectBaseEntity(entity Entity)
	InjectBaseEntitys(entitys interface{})
	GetRuntime() Runtime
}

// QueryBuilder .
type QueryBuilder interface {
	Execute(db *gorm.DB, object interface{}) (e error)
	Order() interface{}
}

var entityPrimary map[reflect.Type][]string = make(map[reflect.Type][]string)
var entityPrimaryLock sync.Mutex

func getEntityPrimary(entity interface{}) []string {
	defer entityPrimaryLock.Unlock()
	entityPrimaryLock.Lock()
	entityType := reflect.TypeOf(entity)
	primarys, ok := entityPrimary[entityType]
	if ok {
		return primarys
	}
	scope := new(gorm.Scope)
	fields := scope.New(entity).PrimaryFields()
	for i := 0; i < len(fields); i++ {
		primarys = append(primarys, fields[i].DBName)
	}
	if len(primarys) == 0 {
		panic("Not found 'gorm:primary_key'")
	}
	entityPrimary[entityType] = primarys
	return primarys
}
