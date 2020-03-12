package general

import "github.com/jinzhu/gorm"

// GORMRepository .
type GORMRepository interface {
	DB() *gorm.DB
	MadeEntity(entity Entity)
	MadeEntitys(entitys interface{})
	GetRuntime() Runtime
}

// QueryBuilder .
type QueryBuilder interface {
	Execute(db *gorm.DB, object interface{}) (e error)
	Order() interface{}
}
