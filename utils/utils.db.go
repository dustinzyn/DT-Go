// 提供数据库连接池的创建
package utils

import (
	"database/sql"
	"errors"
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
)

var dbOnce sync.Once
var db *gorm.DB

type DBConf struct {
	Host         string `yaml:"db_host"`
	Port         int    `yaml:"db_port"`
	User         string `yaml:"db_user"`
	Pwd          string `yaml:"db_pwd"`
	DBName       string `yaml:"db_name"`
	Charset      string `yaml:"charset"`
	MaxOpenConns int    `yaml:"max_open_conns"` // 允许打开的最大连接数
	MaxIdleConns int    `yaml:"max_idle_conns"` // 连接池里的空闲连接数
	Timeout      int    `yaml:"timeout"`        // 连接超时时间 单位毫秒
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	Driver       string `yaml:"driver"`
	Timezone     string `yaml:"timezone"`
	ParseTime    bool   `yaml:"parse_time"`    // 支持把数据库datetime和date类型转换为golang的time.Time类型
	PrintSqlLog  bool   `yaml:"print_sql_log"` // 慢sql时间,单位毫秒,超过这个时间会打印sql
	SlowSqlTime  int    `yaml:"slow_sql_time"` // 是否打印sql, 配合慢sql使用 单位毫秒
}

func NewDBConf() *DBConf {
	return &DBConf{
		Host:         "mariadb-mariadb-cluster.resource.svc.cluster.local",
		Port:         3330,
		User:         "anyshare",
		Pwd:          "eisoo.com123",
		Charset:      "utf8mb4",
		MaxOpenConns: 20,
		MaxIdleConns: 5,
		Timeout:      10000,
		ReadTimeout:  10000,
		WriteTimeout: 10000,
		Driver:       "mysql",
		Timezone:     "Asia/Shanghai",
		ParseTime:    true,
		PrintSqlLog:  true,
		SlowSqlTime:  1000,
	}
}

// ConnectDB return a db conn pool.
func ConnectDB(conf DBConf) (*gorm.DB, error) {
	if conf.DBName == "" {
		return nil, errors.New("Invalid database name")
	}
	dbOnce.Do(func() {
		var err error
		switch conf.Driver {
		case "mysql":
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
	return db, nil
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
