package transaction

/**
数据库事务组件

Created by Dustin.zhu on 2022/11/1.
*/

import (
	"Hive"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func init() {
	hive.Prepare(func(initiator hive.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *GormImpl {
			return &GormImpl{}
		})
	})
}

var _ Transaction = (*GormImpl)(nil)

// Transaction .
type Transaction interface {
	Execute(fun func() error) (err error)
	ExecuteTx(fun func() error, opts *sql.TxOptions) (err error)
}

// GormImpl .
type GormImpl struct {
	hive.Infra
	db *gorm.DB
}

// BeginRequest .
func (t *GormImpl) BeginRequest(worker hive.Worker) {
	t.db = nil
	t.Infra.BeginRequest(worker)
}

// Execute .
func (t *GormImpl) Execute(fun func() error) (err error) {
	return t.execute(fun, nil)
}

// ExecuteTx .
func (t *GormImpl) ExecuteTx(fun func() error, opts *sql.TxOptions) (err error) {
	return t.execute(fun, opts)
}

// execute .
func (t *GormImpl) execute(fun func() error, opts *sql.TxOptions) (err error) {
	if t.db != nil {
		panic("unknown error")
	}

	db := t.SourceDB().(*gorm.DB)
	if opts != nil {
		t.db = db.Begin(opts)
	} else {
		t.db = db.Begin()
	}

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
			if e2.Error != nil {
				err = errors.New(err.Error() + "," + e2.Error.Error())
			}
			return
		}
		err = deferDb.Commit().Error
	}()
	err = fun()
	return
}
