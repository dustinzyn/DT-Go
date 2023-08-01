package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/kataras/iris/v12"
	"gopkg.in/yaml.v3"
)

// Configure
type Configurer interface {
	Configure(obj interface{}, file string, metadata ...interface{}) error
}

var configurer Configurer

// SetConfigurer
func SetConfigurer(confer Configurer) {
	configurer = confer
}

var ProfileENV = "CONFIG_PATH"

// Configure
func Configure(obj interface{}, file string, metadata ...interface{}) (err error) {
	if configurer != nil {
		return configurer.Configure(obj, file, metadata...)
	}
	path := os.Getenv(ProfileENV)
	if path == "" {
		path = "./conf"
		if _, err := os.Stat(path); err != nil {
			path = "./server/conf"
			if _, err := os.Stat(path); err != nil {
				path = ""
			}
		}
	}
	ioStream, err := ioutil.ReadFile(path + "/" + file)
	if err != nil {
		fmt.Printf("Configure readfile error: %v", err)
		return
	}
	err = yaml.Unmarshal(ioStream, obj)
	if err != nil {
		fmt.Printf("Configure decode error: %s", err.Error())
		return
	} else {
		fmt.Printf("Configure decode: %s", path+"/"+file)
	}
	return
}

// Configuration 服务配置
type Configurations struct {
	App   *iris.Configuration  // Application配置
	DB    *DBConfiguration     // Database配置
	Redis *RedisConfiguration  // Redis配置
	MQ    *MQConfiguration     // MQ配置
	DS    *DepSvcConfiguration // 依赖的第三方服务配置
}

// NewConfiguration 初始化默认配置
func NewConfiguration() *Configurations {
	irisCg := iris.DefaultConfiguration()
	dbCg := &DBConfiguration{
		Host:         "mariadb-mariadb-cluster.resource.svc.cluster.local",
		Port:         3330,
		Type:         "mysql",
		User:         "anyshare",
		Pwd:          "eisoo.com123",
		Charset:      "utf8mb4",
		MaxOpenConns: 20,
		MaxIdleConns: 5,
		Timeout:      10000,
		ReadTimeout:  10000,
		WriteTimeout: 10000,
		Driver:       "proton-rds",
		Timezone:     "Asia/Shanghai",
		ParseTime:    true,
		PrintSqlLog:  true,
		SlowSqlTime:  1000,
	}
	redisCg := &RedisConfiguration{
		MaxRetries:         0,
		PoolSize:           2 * runtime.NumCPU(),
		ReadTimeout:        3,
		WriteTimeout:       3,
		IdleTimeout:        300,
		IdleCheckFrequency: 60,
	}
	dsCg := &DepSvcConfiguration{
		UserMgntProtocol:    "http",
		UserMgntHost:        "user-management-private.anyshare.svc.cluster.local",
		UserMgntPort:        "30980",
		HydraPublicProtocol: "http",
		HydraPublicHost:     "hydra-public.anyshare.svc.cluster.local",
		HydraPublicPort:     "4444",
		HydraAdminProtocol:  "http",
		HydraAdminHost:      "hydra-admin.anyshare.svc.cluster.local",
		HydraAdminPort:      "4445",
	}
	mqCg := &MQConfiguration{
		ConnectType:  "nsq",
		ProducerHost: "proton-mq-nsq-nsqd.resource.svc.cluster.local",
		ProducerPort: "4151",
		ConsumerHost: "proton-mq-nsq-nsqlookupd.resource.svc.cluster.local",
		ConsumerPort: "4161",
	}
	return &Configurations{
		DB:    dbCg,
		Redis: redisCg,
		App:   &irisCg,
		DS:    dsCg,
		MQ:    mqCg,
	}
}

type DBConfiguration struct {
	Host         string      `yaml:"db_host"`
	Port         int         `yaml:"db_port"`
	Type         string      `yaml:"db_type"`
	User         string      `yaml:"db_user"`
	Pwd          string      `yaml:"db_pwd"`
	DBName       string      `yaml:"db_name"`
	Charset      string      `yaml:"charset"`
	MaxOpenConns int         `yaml:"max_open_conns"` // 允许打开的最大连接数
	MaxIdleConns int         `yaml:"max_idle_conns"` // 连接池里的空闲连接数
	Timeout      int         `yaml:"timeout"`        // 连接超时时间 单位毫秒
	ReadTimeout  int         `yaml:"read_timeout"`
	WriteTimeout int         `yaml:"write_timeout"`
	Driver       string      `yaml:"driver"`
	Timezone     string      `yaml:"timezone"`
	ParseTime    bool        `yaml:"parse_time"`    // 支持把数据库datetime和date类型转换为golang的time.Time类型
	PrintSqlLog  bool        `yaml:"print_sql_log"` // 慢sql时间,单位毫秒,超过这个时间会打印sql
	SlowSqlTime  int         `yaml:"slow_sql_time"` // 是否打印sql, 配合慢sql使用 单位毫秒
	Other        interface{} `yaml:"Other"`
}

type RedisConfiguration struct {
	ConnectType      string `yaml:"connect_type"`
	MasterGroupName  string `yaml:"master_group_name"`
	SentinelHost     string `yaml:"sentinel_host"`
	SentinelPort     string `yaml:"sentinel_port"`
	SentinelUsername string `yaml:"sentinel_username"`
	SentinelPwd      string `yaml:"sentinel_password"`

	UserName string `yaml:"user_name"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`

	ClusterHosts []string `yaml:"cluster_addrs"`
	ClusterPwd   string   `yaml:"cluster_password"`

	DB                 int `yaml:"db"`
	MaxRetries         int `yaml:"max_retries"`
	PoolSize           int `yaml:"pool_size"`
	ReadTimeout        int `yaml:"read_timeout"`
	WriteTimeout       int `yaml:"write_timeout"`
	IdleTimeout        int `yaml:"idle_timeout"`
	IdleCheckFrequency int `yaml:"idle_check_frequency"`
	MaxConnAge         int `yaml:"max_conn_age"`
	PoolTimeout        int `yaml:"pool_timeout"`

	Other interface{} `yaml:"Other"`
}

type MQConfiguration struct {
	ConnectType  string      `yaml:"connect_type"`
	ProducerHost string      `yaml:"producer_host"`
	ProducerPort string      `yaml:"producer_port"`
	ConsumerHost string      `yaml:"consumer_host"`
	ConsumerPort string      `yaml:"consumer_port"`
	Other        interface{} `yaml:"Other"`
}

// DepSvcConfiguration 公共的依赖服务配置
type DepSvcConfiguration struct {
	UserMgntProtocol    string      `yaml:"user_management_private_protocol"`
	UserMgntHost        string      `yaml:"user_management_private_host"`
	UserMgntPort        string      `yaml:"user_management_private_port"`
	HydraPublicProtocol string      `yaml:"hydra_public_protocol"`
	HydraPublicHost     string      `yaml:"hydra_public_host"`
	HydraPublicPort     string      `yaml:"hydra_public_port"`
	HydraAdminProtocol  string      `yaml:"hydra_admin_protocol"`
	HydraAdminHost      string      `yaml:"hydra_admin_host"`
	HydraAdminPort      string      `yaml:"hydra_admin_port"`
	Other               interface{} `yaml:"Other"`
}

func (cg *Configurations) ConfigureApp(file string) {
	Configure(cg.App, file)
}

func (cg *Configurations) ConfigureDB(file string) {
	Configure(cg.DB, file)
}

func (cg *Configurations) ConfigureRedis(file string) {
	Configure(cg.Redis, file)
}

func (cg *Configurations) ConfigureMQ(file string) {
	Configure(cg.MQ, file)
}

func (cg *Configurations) ConfigureDS(file string) {
	Configure(cg.DS, file)
}
