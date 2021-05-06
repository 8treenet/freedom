package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
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
	Execute(fun func() error, opts ...*sql.TxOptions) (e error)
}

var _ Transaction = (*GormImpl)(nil)

//GormImpl .
type GormImpl struct {
	freedom.Infra
}

// BeginRequest .
func (t *GormImpl) BeginRequest(worker freedom.Worker) {
	t.Infra.BeginRequest(worker)
}

// Execute .
func (t *GormImpl) Execute(fun func() error, opts ...*sql.TxOptions) (e error) {
	return t.execute(t.Worker().Context(), fun, opts...)
}

// execute .
func (t *GormImpl) execute(ctx context.Context, fun func() error, opts ...*sql.TxOptions) (e error) {
	var db *gorm.DB
	if err := t.FetchOnlyDB(&db); err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) (e error) {
		t.Worker().Store().Set(freedom.TransactionKey, tx)
		defer func() {
			if perr := recover(); perr != nil {
				e = errors.New(fmt.Sprint(perr))
			}
			t.Worker().Store().Remove(freedom.TransactionKey)
		}()

		e = fun()
		return
	}, opts...)
}
