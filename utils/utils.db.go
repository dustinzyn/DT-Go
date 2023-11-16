// 提供数据库连接池的创建
package utils

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	dt "DT-Go"
	"DT-Go/config"
)

var (
	dbOnce   sync.Once
	rwdbOnce sync.Once
	// db gorm数据库连接池对象
	db *gorm.DB
)

// ConnectDB return a db conn pool.
func ConnectDB(conf *config.DBConfiguration) *gorm.DB {
	dbOnce.Do(func() {
		var err error
		switch conf.Driver {
		case "mysql":
			if conf.DBName == "" {
				panic(fmt.Errorf("Invalid database name"))
			}
			dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%s&loc=Local&timeout=%dms",
				conf.User, conf.Pwd, conf.Host, conf.Port, conf.DBName, conf.Charset, strconv.FormatBool(conf.ParseTime), conf.Timeout)
			ormconf := gorm.Config{}
			if conf.PrintSqlLog {
				loggerNew := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
					SlowThreshold:             time.Duration(conf.SlowSqlTime) * time.Millisecond, //慢SQL阈值
					LogLevel:                  logger.Warn,
					Colorful:                  false, // 彩色打印开启
					IgnoreRecordNotFoundError: true,
				})
				ormconf.Logger = loggerNew
			}
			db, err = gorm.Open(mysql.Open(dsn), &ormconf)
			if err != nil {
				panic(err)
			}
			var opt *sql.DB
			opt, err = db.DB()
			if err != nil {
				panic(err)
			}
			opt.SetMaxIdleConns(conf.MaxIdleConns)
			opt.SetMaxOpenConns(conf.MaxOpenConns)
		case "sqlite3":
			dsn := os.Getenv("DB_URL")
			db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
			if err != nil {
				panic(err)
			}
		}
		dt.Logger().Infof("connect database success...")
	})
	return db
}

// DisconnectDB .
func DisconnectDB() error {
	if db != nil {
		opt, _ := db.DB()
		err := opt.Close()
		db = nil
		return err
	}
	return nil
}
