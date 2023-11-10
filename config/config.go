package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sync"

	sentinel "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/DT-Go/infra/rate/sentinel/api"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/DT-Go/infra/rate/sentinel/core/flow"
	"devops.aishu.cn/AISHUDevOps/ONE-Architecture/_git/proton-rds-sdk-go/sqlx"
	"github.com/kataras/iris/v12"
	"gopkg.in/yaml.v3"
)

const (
	// ProfileENV 配置文件所在目录的环境变量
	ProfileENV = "CONFIG_PATH"
)

var (
	cgOnce sync.Once
	// configurer 配置器
	configurer Configurer
	// configuration 配置
	configuration *Configurations
)

// Configure
type Configurer interface {
	Configure(obj interface{}, file string, metadata ...interface{}) error
}

// SetConfigurer
func SetConfigurer(confer Configurer) {
	configurer = confer
}

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
		fmt.Printf("Configure readfile error: %v\n", err)
		return
	}
	err = yaml.Unmarshal(ioStream, obj)
	if err != nil {
		fmt.Printf("Configure decode error: %s\n", err.Error())
		return
	} else {
		fmt.Printf("Configure decode: %s\n", path+"/"+file)
	}
	return
}

// Configuration 服务配置
type Configurations struct {
	App      *iris.Configuration      // Application配置
	DB       *DBConfiguration         // Database配置
	RWDB     *sqlx.DBConfig           // Database读写分离配置
	Redis    *RedisConfiguration      // Redis配置
	MQ       *MQConfiguration         // MQ配置
	DS       *DepSvcConfiguration     // 依赖的第三方服务配置
	RateRule []*RateRuleConfiguration // 限流配置
}

// NewConfiguration 初始化默认配置
func NewConfiguration() *Configurations {
	cgOnce.Do(func() {
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
		rwdbCg := &sqlx.DBConfig{
			Host:             "mariadb-mariadb-cluster.resource.svc.cluster.local",
			Port:             3330,
			HostRead:         "mariadb-mariadb-cluster.resource.svc.cluster.local",
			PortRead:         3330,
			User:             "anyshare",
			Password:         "eisoo.com123",
			Charset:          "utf8mb4",
			MaxOpenConns:     20,
			Timeout:          10000,
			ReadTimeout:      10000,
			WriteTimeout:     10000,
			MaxOpenReadConns: 20,
		}
		redisCg := &RedisConfiguration{
			UserName:           "root",
			Password:           "eisoo.com123",
			DB:                 10,
			MaxRetries:         10,
			PoolSize:           2 * runtime.NumCPU(),
			ReadTimeout:        3,
			WriteTimeout:       3,
			IdleTimeout:        300,
			IdleCheckFrequency: 60,
			MaxConnAge:         300,
			PoolTimeout:        8,
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
		configuration = &Configurations{
			DB:       dbCg,
			RWDB:     rwdbCg,
			Redis:    redisCg,
			App:      &irisCg,
			DS:       dsCg,
			MQ:       mqCg,
			RateRule: make([]*RateRuleConfiguration, 0),
		}
	})
	return configuration
}

type DBConfiguration struct {
	Host         string      `yaml:"db_host"`
	Port         int         `yaml:"db_port"`
	Type         string      `yaml:"db_type"` // 类型 mysql dm8
	User         string      `yaml:"user_name"`
	Pwd          string      `yaml:"user_pwd"`
	DBName       string      `yaml:"db_name"`
	Charset      string      `yaml:"db_charset"`
	MaxOpenConns int         `yaml:"max_open_conns"` // 允许打开的最大连接数
	MaxIdleConns int         `yaml:"max_idle_conns"` // 连接池里的空闲连接数
	Timeout      int         `yaml:"timeout"`        // 连接超时时间 单位毫秒
	ReadTimeout  int         `yaml:"read_timeout"`   // 读超时时间
	WriteTimeout int         `yaml:"write_timeout"`  // 写超时时间
	Driver       string      `yaml:"driver"`         // 驱动 proton-rds: proton提供的 sqlite3: 单元测试用
	Timezone     string      `yaml:"timezone"`
	ParseTime    bool        `yaml:"parse_time"`    // 支持把数据库datetime和date类型转换为golang的time.Time类型
	PrintSqlLog  bool        `yaml:"print_sql_log"` // 慢sql时间,单位毫秒,超过这个时间会打印sql
	SlowSqlTime  int         `yaml:"slow_sql_time"` // 是否打印sql, 配合慢sql使用 单位毫秒
	Other        interface{} `yaml:"Other"`
}

type RedisConfiguration struct {
	UserName string `yaml:"user_name"`
	Password string `yaml:"password"`

	ConnectType string `yaml:"connect_type"` // 部署方式 sentinel:哨兵模式 master-slave:主从模式 cluster:集群模式 standalone:单机模式

	// sentinel
	MasterGroupName  string `yaml:"master_group_name"`
	SentinelHost     string `yaml:"sentinel_host"`
	SentinelPort     string `yaml:"sentinel_port"`
	SentinelUsername string `yaml:"sentinel_username"`
	SentinelPwd      string `yaml:"sentinel_password"`

	// standalone
	Host string `yaml:"host"`
	Port string `yaml:"port"`

	// master-slave
	MasterHost string `yaml:"master_host"`
	MasterPort string `yaml:"master_port"`
	SlaveHost  string `yaml:"slave_host"`
	SlavePort  string `yaml:"slave_port"`

	// cluster
	ClusterHosts []string `yaml:"cluster_addrs"`
	ClusterPwd   string   `yaml:"cluster_password"`

	DB                 int `yaml:"db"`                   // 数据库 默认第10个
	MaxRetries         int `yaml:"max_retries"`          // 最大重试次数
	PoolSize           int `yaml:"pool_size"`            // 连接池大小
	ReadTimeout        int `yaml:"read_timeout"`         // 读取超时时间 默认3秒
	WriteTimeout       int `yaml:"write_timeout"`        // 写入超时时间 默认3秒
	IdleTimeout        int `yaml:"idle_timeout"`         // 连接空闲时间 默认300秒
	IdleCheckFrequency int `yaml:"idle_check_frequency"` // 检测死连接并清理 默认60秒
	MaxConnAge         int `yaml:"max_conn_age"`         // 连接最长时间 默认300秒
	PoolTimeout        int `yaml:"pool_timeout"`         // 如果连接池已满 等待可用连接的时间 默认8秒

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

// RateRuleConfiguration 限流配置
// Rule describes the strategy of flow control, the flow control strategy is based on QPS statistic metric
// StatIntervalInMs 和 Threshold 这两个字段，这两个字段决定了流量控制器的灵敏度。
// 以 Direct + Reject 的流控策略为例，流量控制器的行为就是在 StatIntervalInMs 周期内，
// 允许的最大请求数量是Threshold。比如如果 StatIntervalInMs 是 10000，Threshold 是10000，
// 那么流量控制器的行为就是控制该资源10s内运行最多10000次访问。
type RateRuleConfiguration struct {
	// Resource represents the resource name.
	Resource string `yaml:"resource" json:"resource"`

	// TokenCalculateStrategy means the control strategy of the flow controller;
	// Reject indicates a direct rejection when the threshold is exceeded,
	// and Throttling indicates a uniform queue speed. The token calculation strategy for the current traffic controller.
	// Direct means directly using the field Threshold as the threshold
	TokenCalculateStrategy string `json:"token_calculate_strategy"`

	// ControlBehavior means the control strategy of the flow controller;
	// Reject indicates a direct rejection when the threshold is exceeded,
	// and Throttling indicates a uniform queue speed.
	ControlBehavior string `yaml:"control_behavior" json:"control_behavior"`

	// Threshold means the threshold during StatIntervalInMs
	// If StatIntervalInMs is 1000(1 second), Threshold means QPS
	Threshold float64 `yaml:"threshold" json:"threshold"`

	// MaxQueueingTimeMs only takes effect when ControlBehavior is Throttling.
	// When MaxQueueingTimeMs is 0, it means Throttling only controls interval of requests,
	// and requests exceeding the threshold will be rejected directly.
	MaxQueueingTimeMs int `yaml:"max_queueing_time_ms" json:"max_queueing_time_ms"`

	// StatIntervalInMs indicates the statistic interval and it's the optional setting for flow Rule.
	// If user doesn't set StatIntervalInMs, that means using default metric statistic of resource.
	// If the StatIntervalInMs user specifies can not reuse the global statistic of resource,
	// sentinel will generate independent statistic structure for this rule.
	StatIntervalInMs int `yaml:"stat_interval_in_ms" json:"stat_interval_in_ms"`
}

func (cg *Configurations) ConfigureApp(file string) {
	Configure(cg.App, file)
}

func (cg *Configurations) ConfigureDB(file string) {
	Configure(cg.DB, file)
}

func (cg *Configurations) ConfigureRWDB(file string) {
	Configure(cg.RWDB, file)
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

func (cg *Configurations) ConfigureRateRule(file string) {
	Configure(&cg.RateRule, file)
	err := sentinel.InitDefault()
	if err != nil {
		panic(err)
	}
	// 添加规则
	flowRules := make([]*flow.Rule, 0)
	for _, rule := range cg.RateRule {
		behavior := flow.Reject
		switch rule.ControlBehavior {
		case "Reject":
			behavior = flow.Reject
		case "Throttling":
			behavior = flow.Throttling
		}

		flowRule := flow.Rule{
			Resource:               rule.Resource,
			Threshold:              rule.Threshold,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        behavior,
			StatIntervalInMs:       uint32(rule.StatIntervalInMs),
		}
		if behavior == flow.Throttling {
			flowRule.MaxQueueingTimeMs = uint32(rule.MaxQueueingTimeMs)
		}
		flowRules = append(flowRules, &flowRule)
	}
	if _, err := flow.LoadRules(flowRules); err != nil {
		fmt.Println("Failed to load rules:", err)
	}
}
