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
		initiator.BindInfra(false, func() *GormImpl {
			return &GormImpl{}
		})
	})
}

// Transaction .
type Transaction interface {
	Execute(fun func() error) (e error)
	ExecuteTx(ctx context.Context, fun func() error, opts *sql.TxOptions) (e error)
}

var _ Transaction = (*GormImpl)(nil)

//GormImpl .
type GormImpl struct {
	freedom.Infra
	db *gorm.DB
}

// BeginRequest .
func (t *GormImpl) BeginRequest(worker freedom.Worker) {
	t.db = nil
	t.Infra.BeginRequest(worker)
}

// Execute .
func (t *GormImpl) Execute(fun func() error) (e error) {
	return t.execute(nil, fun, nil)
}

// ExecuteTx .
func (t *GormImpl) ExecuteTx(ctx context.Context, fun func() error, opts *sql.TxOptions) (e error) {
	return t.execute(ctx, fun, opts)
}

// execute .
func (t *GormImpl) execute(ctx context.Context, fun func() error, opts *sql.TxOptions) (e error) {
	if t.db != nil {
		panic("unknown error")
	}

	db := t.SourceDB().(*gorm.DB)
	if ctx != nil && opts != nil {
		t.db = db.BeginTx(ctx, opts)
	} else {
		t.db = db.Begin()
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
