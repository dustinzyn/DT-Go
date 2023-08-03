package transaction

/**
数据库事务组件

Created by Dustin.zhu on 2022/11/1.
*/

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	hive "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	"devops.aishu.cn/AISHUDevOps/ONE-Architecture/_git/proton-rds-sdk-go/sqlx"
)

//go:generate mockgen -package mock_infra -source transaction.go -destination ./mock/transaction_mock.go

func init() {
	hive.Prepare(func(initiator hive.Initiator) {
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
	hive.Infra
	db *sql.Tx
}

// BeginRequest .
func (t *SqlDBImpl) BeginRequest(worker hive.Worker) {
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

	db := t.SourceDB().(*sqlx.DB)
	if opts != nil {
		t.db, err = db.BeginTx(context.Background(), opts)
		if err != nil {
			return
		}
	} else {
		t.db, err = db.Begin()
		if err != nil {
			return
		}
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
			if e2 != nil {
				err = errors.New(err.Error() + "," + e2.Error())
			}
			return
		}
		err = deferDb.Commit()
	}()
	err = fun()
	return
}
