package transaction

import (
	"errors"
	"fmt"

	"github.com/8treenet/freedom"
	"github.com/jinzhu/gorm"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindInfra(false, func() *Transaction {
			return &Transaction{}
		})
	})
}

// Transaction .
type Transaction struct {
	freedom.Infra
	db *gorm.DB
}

// BeginRequest
func (t *Transaction) BeginRequest(rt freedom.Runtime) {
	t.db = nil
	t.Infra.BeginRequest(rt)
}

// Execute Execute local transaction.
func (t *Transaction) Execute(fun func() error, selectDBName ...string) (e error) {
	if t.db != nil {
		panic("unknown error")
	}
	name := ""
	if len(selectDBName) > 0 {
		name = selectDBName[0]
	}
	t.db = t.DB(selectDBName...).Begin()
	t.Runtime.Store().Set("freedom_local_transaction_db", map[string]interface{}{
		"db":   t.db,
		"name": name,
	})

	defer func() {
		if perr := recover(); perr != nil {
			t.db.Rollback()
			t.db = nil
			e = errors.New(fmt.Sprint(perr))
			t.Runtime.Store().Remove("freedom_local_transaction_db")
			return
		}

		deferdb := t.db
		t.Runtime.Store().Remove("freedom_local_transaction_db")
		t.db = nil
		if e != nil {
			e2 := deferdb.Rollback()
			if e2.Error != nil {
				e = errors.New(e.Error() + "," + e2.Error.Error())
			}
			return
		}
		e = deferdb.Commit().Error
	}()
	e = fun()
	return
}
