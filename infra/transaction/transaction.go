package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/8treenet/freedom"
	"github.com/jinzhu/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindInfra(false, func() *TransactionImpl {
			return &TransactionImpl{}
		})
	})
}

type Transaction interface {
	Execute(fun func() error) (e error)
	ExecuteTx(fun func() error, ctx context.Context, opts *sql.TxOptions) (e error)
}

// TransactionImpl .
type TransactionImpl struct {
	freedom.Infra
	db *gorm.DB
}

// BeginRequest
func (t *TransactionImpl) BeginRequest(worker freedom.Worker) {
	t.db = nil
	t.Infra.BeginRequest(worker)
}

func (t *TransactionImpl) Execute(fun func() error) (e error) {
	return t.execute(fun, nil, nil)
}

// Execute Execute local transaction.
func (t *TransactionImpl) ExecuteTx(fun func() error, ctx context.Context, opts *sql.TxOptions) (e error) {
	return t.execute(fun, ctx, opts)
}

// Execute Execute local transaction.
func (t *TransactionImpl) execute(fun func() error, ctx context.Context, opts *sql.TxOptions) (e error) {
	if t.db != nil {
		panic("unknown error")
	}
	if ctx != nil && opts != nil {
		t.db = t.DB().BeginTx(ctx, opts)
	} else {
		t.db = t.DB().Begin()
	}

	t.Worker.Store().Set("freedom_local_transaction_db", t.db)

	defer func() {
		if perr := recover(); perr != nil {
			t.db.Rollback()
			t.db = nil
			e = errors.New(fmt.Sprint(perr))
			t.Worker.Store().Remove("freedom_local_transaction_db")
			return
		}

		deferdb := t.db
		t.Worker.Store().Remove("freedom_local_transaction_db")
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
