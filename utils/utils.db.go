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

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/config"
	dm "devops.aishu.cn/AISHUDevOps/ONE-Architecture/_git/proton_dm_dialect_go"

	_ "devops.aishu.cn/AISHUDevOps/ONE-Architecture/_git/proton-rds-sdk-go/driver" // 注册数据库驱动
)

var dbOnce sync.Once
var db *gorm.DB

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
	})
	return db
}

// ConnProtonRDS return a db conn pool.
func ConnProtonRDS(conf *config.DBConfiguration) *gorm.DB {
	dbOnce.Do(func() {
		var err error
		switch conf.Driver {
		case "proton-rds":
			if conf.DBName == "" {
				panic(fmt.Errorf("Invalid database name."))
			}
			dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%s&loc=Local&timeout=%dms",
				conf.User, conf.Pwd, conf.Host, conf.Port, conf.DBName, conf.Charset, strconv.FormatBool(conf.ParseTime), conf.Timeout)
			ormconf := gorm.Config{SkipDefaultTransaction: true}
			if conf.PrintSqlLog {
				loggerNew := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
					SlowThreshold:             time.Duration(conf.SlowSqlTime) * time.Millisecond, //慢SQL阈值
					LogLevel:                  logger.Warn,
					Colorful:                  false, // 彩色打印开启
					IgnoreRecordNotFoundError: true,
				})
				ormconf.Logger = loggerNew
			}

			operation, err := sql.Open(conf.Driver, dsn)
			if err != nil {
				panic(err)
			}
			var dialector gorm.Dialector
			if conf.Type == "DM8" {
				dialector = dm.New(dm.Config{Conn: operation})
			} else {
				// mysql mariadb tidb
				dialector = mysql.New(mysql.Config{Conn: operation})
			}
			db, err = gorm.Open(dialector, &ormconf)
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
		default:
			panic(fmt.Errorf("Invalid database driver."))
		}
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
