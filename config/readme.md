# 框架配置说明

## 配置项
```golang
type Configurations struct {
	App   *iris.Configuration  // Application配置
	DB    *DBConfiguration     // Database配置
	Redis *RedisConfiguration  // Redis配置
	MQ    *MQConfiguration     // MQ配置
	DS    *DepSvcConfiguration // 依赖的第三方服务配置
}
```
* App: Appliction的配置，提供了默认配置项，详细参考*iris.Configuration*，可增加自定义的配置项 Other
* DB: 数据库配置，可增加自定义配置项 Other
* Redis: Redis配置，可增加自定义配置项 Other
* MQ: 消息代理配置，可增加自定义配置项 Other
* DS: 第三方依赖服务配置，可增加自定义配置项 Other

## 如何使用
### 1.获取配置类
``` golang
Cfg = hive.NewConfiguration()
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