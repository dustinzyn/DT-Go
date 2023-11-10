# 框架配置说明

## 配置项
```golang
type Configurations struct {
	App      *iris.Configuration      // Application配置
	DB       *DBConfiguration         // Database配置
	RWDB     *sqlx.DBConfig           // Database读写分离配置
	Redis    *RedisConfiguration      // Redis配置
	MQ       *MQConfiguration         // MQ配置
	DS       *DepSvcConfiguration     // 依赖的第三方服务配置
	RateRule []*RateRuleConfiguration // 限流配置
}
```
* App: Appliction的配置，提供了默认配置项，详细参考*iris.Configuration*，可增加自定义的配置项 Other
* DB: 数据库配置，可增加自定义配置项 Other
* Redis: Redis配置，可增加自定义配置项 Other
* MQ: 消息代理配置，可增加自定义配置项 Other
* DS: 第三方依赖服务配置，可增加自定义配置项 Other
* RateRule：限流配置

## 如何使用
### 1.获取配置类
``` golang
Cfg = dt.NewConfiguration()
```
### 2.加载配置
* 若指定了环境变量 CONFIG_PATH，则配置文件的读取路径为 CONFIG_PATH/file.yaml，适用于生产环境。
* 未指定环境变量 CONFIG_PATH，则配置文件的读取路径为项目目录下conf/file.yaml，读取不到，会读取server/conf/file.yaml，适用于本地调试环境
``` golang
// 加载app配置
Cfg.ConfigureApp("app.yaml")
// 加载db配置
Cfg.ConfigureDB("db.yaml")
// 加载redis配置
Cfg.ConfigureRedis("redis.yaml")
// 加载mq配置
Cfg.ConfigureMQ("mq.yaml")
// 加载依赖服务配置
Cfg.ConfigureDS("depsvc.yaml")
// 加载限流配置
Cfg.ConfigureRateRule("raterule.yaml")
```
### 3.如何增加自定义配置
#### 3.1 增加Application自定义配置
增加一个服务名称的配置项，app.yaml写法如下
``` yaml
Other:
  service_name: hivecore
```
加载app.yaml配置
``` golang
// 加载app配置
Cfg.ConfigureApp("app.yaml")

// 使用
svcName := Cfg.APP.Other["service_name"]
```
#### 3.1 增加第三方依赖服务配置
depsvc.yaml如下
``` yaml
user_management_private_host: user-management-private.anyshare.svc.cluster.local
user_management_private_port: 30980
user_management_private_protocol: http
hydra_public_protocol: http
hydra_public_host: hydra-public.anyshare.svc.cluster.local
hydra_public_port: 4444
hydra_admin_protocol: http
hydra_admin_host: hydra-admin.anyshare.svc.cluster.local
hydra_admin_port: 4445
Other:
  hivecore_private_host: 127.0.0.1
  hivecore_private_port: 8879
  hivecore_private_protocol: http
```
加载depsvc.yaml配置
``` golang
// 加载app配置
Cfg.ConfigureApp("depsvc.yaml")

// 使用
host := Cfg.DS.Other["hivecore_private_host"]
```

### 4.限流配置常见场景示例
#### 4.1 基于QPS对某个API的资源限流
基于对某个资源访问的QPS来做流控，这个是最常见的场景。

``` yaml
- resource: GET:/api/DT-Gocore/v1/user/{param1:string}
  control_behavior: Throttling
  threshold: 100
  stat_interval_in_ms: 1000
  max_queueing_time_ms: 200
- resource: POST:/api/DT-Gocore/v1/user
  control_behavior: Reject
  threshold: 50
  stat_interval_in_ms: 1000
  max_queueing_time_ms: 200
```
上面sample中的5个字段是必填的。其中StatIntervalInMs必须是1000，表示统计周期是1s，那么Threshold所配置的值也就是QPS的阈值。

这个示例里配置了两条规则，
* 第一条规则表示对“/api/DT-Gocore/v1/user/{param1:string}” 这个API的GET请求限流，每秒允许通过100个请求，超出的请求匀速排队通过（Throttling），上面 Threshold 是 100，Sentinel 默认使用1s作为控制周期，表示1秒内10个请求匀速排队，所以排队时间就是 1000ms/100 = 10ms；
* 第二条规则表示对“/api/DT-Gocore/v1/user” 这个API的POST请求限流，每秒允许通过50个请求，超出的请求直接拒绝，会返回429的状态码

注意：MaxQueueingTimeMs 设为 0 时代表不允许排队，只控制请求时间间隔，多余的请求将会直接拒绝。

### 4.2 基于一定统计间隔时间来控制总的请求数

这个场景就是想在一定统计周期内控制请求的总量。
``` yaml
- resource: POST:/api/DT-Gocore/v1/user
  control_behavior: Reject
  threshold: 1000
  stat_interval_in_ms: 10000
```
这个规则表示对“/api/DT-Gocore/v1/user” 这个API的POST请求限流，在10s的时间内，允许通过1000个请求

### 4.3 毫秒级别流控

针对一些流量在毫秒级别波动非常大的场景(类似于脉冲)，建议StatIntervalInMs的配置在毫秒级别，建议配置的值为100ms的倍数。这种相当于缩小了统计周期，将QPS的周期缩小了10倍，控制周期降低到了100ms。这种配置能够很好的应对脉冲流量，保障系统稳定性。
``` yaml
- resource: POST:/api/DT-Gocore/v1/user
  control_behavior: Reject
  threshold: 60
  stat_interval_in_ms: 100
```
这个规则表示对“/api/DT-Gocore/v1/user” 这个API的POST请求限流，限制了100ms的阈值是60，实际的QPS大概是600。这个配置一般用于处理脉冲流量，所以某些时间段的脉冲流量如果较大，会拒绝掉很多流量。

如果既想控制流量曲线，又想无损，一般做法是通过匀速排队的控制策略，平滑掉流量。可以参考如下配置
``` yaml
- resource: POST:/api/DT-Gocore/v1/user
  control_behavior: Throttling
  threshold: 60
  stat_interval_in_ms: 100
```
