package transaction

/**
数据库事务组件

Created by Dustin.zhu on 2022/11/1.
*/

import (
	"database/sql"
	"errors"
	"fmt"

	dt "DT-Go"

	"gorm.io/gorm"
)

//go:generate mockgen -package mock_infra -source transaction.go -destination ./mock/transaction_mock.go

func init() {
	dt.Prepare(func(initiator dt.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *SqlDBImpl {
			return &SqlDBImpl{}
		})
	})
}

var _ Transaction = (*SqlDBImpl)(nil)

// Transaction .
type Transaction interface {
	Execute(fun func() error) (err error)
	ExecuteTx(fun func() error, opts *sql.TxOptions) (err error)
}

// SqlDBImpl .
type SqlDBImpl struct {
	dt.Infra
	db *gorm.DB
}

// BeginRequest .
func (t *SqlDBImpl) BeginRequest(worker dt.Worker) {
	t.db = nil
	t.Infra.BeginRequest(worker)
}

// Execute .
func (t *SqlDBImpl) Execute(fun func() error) (err error) {
	return t.execute(fun, nil)
}

// ExecuteTx .
func (t *SqlDBImpl) ExecuteTx(fun func() error, opts *sql.TxOptions) (err error) {
	return t.execute(fun, opts)
}

// execute .
func (t *SqlDBImpl) execute(fun func() error, opts *sql.TxOptions) (err error) {
	if t.db != nil {
		panic("unknown error")
	}

	db := t.SourceDB().(*gorm.DB)
	t.db = db.Begin(opts)

	t.Worker().Store().Set("local_transaction_db", t.db)

	defer func() {
		if perr := recover(); perr != nil {
			t.db.Rollback()
			t.db = nil
			err = errors.New(fmt.Sprint(perr))
			t.Worker().Store().Remove("local_transaction_db")
			return
		}
		deferDb := t.db
		t.Worker().Store().Remove("local_transaction_db")
		t.db = nil
		if err != nil {
			e2 := deferDb.Rollback()
			if e2 != nil {
				err = errors.New(err.Error() + "," + e2.Error.Error())
			}
			return
		}
		deferDb.Commit()
	}()
	err = fun()
	return
}
