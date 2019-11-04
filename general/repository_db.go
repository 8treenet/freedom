package general

import (
	"errors"
	"fmt"

	"github.com/8treenet/gcache"

	"github.com/jinzhu/gorm"
)

// RepositoryDB .
type RepositoryDB struct {
	Runtime       Runtime
	transactionDB *gorm.DB
}

// BeginRequest .
func (repo *RepositoryDB) BeginRequest(rt Runtime) {
	repo.Runtime = rt
}

// DB .
func (repo *RepositoryDB) DB() (db *gorm.DB) {
	if repo.transactionDB != nil {
		return repo.transactionDB
	}

	return globalApp.Database.db
}

// DBCache .
func (repo *RepositoryDB) DBCache() gcache.Plugin {
	return globalApp.Database.cache
}

// Transaction .
func (repo *RepositoryDB) Transaction(fun func() error) (e error) {
	repo.transactionDB = globalApp.Database.db.Begin()

	defer func() {
		if perr := recover(); perr != nil {
			repo.transactionDB.Rollback()
			repo.transactionDB = nil
			e = errors.New(fmt.Sprint(perr))
			return
		}

		deferdb := repo.transactionDB
		repo.transactionDB = nil
		if e != nil {
			e = deferdb.Rollback().Error
			return
		}
		e = deferdb.Commit().Error
	}()

	e = fun()
	return
}

// NewDescOrder .
func (repo *RepositoryDB) NewDescOrder(orderField string, orderFields ...string) *Order {
	fields := []string{orderField}
	fields = append(fields, orderFields...)
	return &Order{
		orderFields: fields,
		order:       "desc",
	}
}

// NewAscOrder .
func (repo *RepositoryDB) NewAscOrder(orderField string, orderFields ...string) *Order {
	fields := []string{orderField}
	fields = append(fields, orderFields...)
	return &Order{
		orderFields: fields,
		order:       "asc",
	}
}
