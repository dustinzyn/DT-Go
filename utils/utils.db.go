// 提供数据库连接池的创建
package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBConf struct {
	Host         string `yaml:"db_host"`
	Port         int    `yaml:"db_port"`
	User         string `yaml:"db_user"`
	Pwd          string `yaml:"db_pwd"`
	DBName       string `yaml:"db_name"`
	Charset      string `yaml:"charset"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	Timeout      int    `yaml:"timeout"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	Driver       string `yaml:"driver"`
	Timezone     string `yaml:"timezone"`
	ParseTime    bool   `yaml:"parse_time"`
	PrintSqlLog  bool   `yaml:"print_sql_log"`
	SlowSqlTime  string `yaml:"slow_sql_time"`
}

// ConnectDB return a db conn pool.
func ConnectDB(conf DBConf) (db *gorm.DB, err error) {
	user := "anyshare"
	if conf.User != "" {
		user = conf.User
	}
	pwd := "eisoo.com123"
	if conf.Pwd != "" {
		pwd = conf.Pwd
	}
	host := "mariadb-mariadb-cluster.resource.svc.cluster.local"
	if conf.Host != "" {
		host = conf.Host
	}
	port := 3330
	if conf.Port != 0 {
		port = conf.Port
	}
	if conf.DBName == "" {
		return nil, errors.New("Invalid database name")
	}
	charset := "utf8mb4"
	if conf.Charset != "" {
		charset = conf.Charset
	}
	// 支持把数据库datetime和date类型转换为golang的time.Time类型
	parseTime := true
	if !conf.ParseTime {
		parseTime = false
	}
	// 慢sql时间,单位毫秒,超过这个时间会打印sql
	slowSqlTime := "1000ms"
	if conf.SlowSqlTime != "" {
		slowSqlTime = conf.SlowSqlTime
	}
	// 是否打印sql, 配合慢sql使用
	printSqlLog := true
	if !conf.PrintSqlLog {
		printSqlLog = conf.PrintSqlLog
	}
	// 连接池里的空闲连接数
	maxIdleConns := 10
	if conf.MaxIdleConns != 0 {
		maxIdleConns = conf.MaxIdleConns
	}
	// 允许打开的最大连接数
	maxOpenConns := 20
	if conf.MaxOpenConns != 0 {
		maxOpenConns = conf.MaxOpenConns
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%s&loc=Local",
		user, pwd, host, port, conf.DBName, charset, strconv.FormatBool(parseTime))
	ormconf := gorm.Config{}
	if printSqlLog {
		slowTime, err := time.ParseDuration(slowSqlTime)
		if err != nil {
			return nil, err
		}
		loggerNew := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             slowTime, //慢SQL阈值
			LogLevel:                  logger.Warn,
			Colorful:                  false, // 彩色打印开启
			IgnoreRecordNotFoundError: true,
		})
		ormconf.Logger = loggerNew
	}
	db, err = gorm.Open(mysql.Open(dsn), &ormconf)
	if err != nil {
		return
	}
	opt, err := db.DB()
	opt.SetMaxIdleConns(maxIdleConns)
	opt.SetMaxOpenConns(maxOpenConns)
	return
}
