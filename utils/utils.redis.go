// 提供redis连接池的创建
package utils

import (
	"context"
	"fmt"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/config"
	redis "github.com/go-redis/redis/v8"
)

// ConnectRedis return a redis client. If not connected,
// it will automatically reconnect until connected.
func ConnectRedis(conf config.RedisConfiguration) (client redis.Cmdable) {
	ctx := context.Background()

	switch conf.ConnectType {
	case "master-slave":
		for {
			client = masterSlave(conf)
			if err := client.Ping(ctx).Err(); err != nil {
				time.Sleep(time.Duration(3) * time.Second)
			} else {
				break
			}
		}
	case "standalone":
		for {
			client = standalone(conf)
			if err := client.Ping(ctx).Err(); err != nil {
				time.Sleep(time.Duration(3) * time.Second)
			} else {
				break
			}
		}
	case "sentinel":
		for {
			client = sentinel(conf)
			if err := client.Ping(ctx).Err(); err != nil {
				time.Sleep(time.Duration(3) * time.Second)
			} else {
				break
			}
		}
	case "cluster":
		for {
			client = cluster(conf)
			if err := client.Ping(ctx).Err(); err != nil {
				time.Sleep(time.Duration(3) * time.Second)
			} else {
				break
			}
		}
	}
	return
}

// masterSlave 主从模式
func masterSlave(conf config.RedisConfiguration) *redis.Client {
	if conf.MasterHost == "" {
		conf.MasterHost = "proton-redis-proton-redis.resource.svc.cluster.local"
	}
	if conf.MasterPort == "" {
		conf.Port = "6379"
	}
	opt := &redis.Options{
		Addr:               conf.MasterHost + ":" + conf.MasterPort,
		Password:           conf.Password,
		DB:                 conf.DB,
		MaxRetries:         conf.MaxRetries,
		PoolSize:           conf.PoolSize,
		ReadTimeout:        time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout:       time.Duration(conf.WriteTimeout) * time.Second,
		IdleTimeout:        time.Duration(conf.IdleTimeout) * time.Second,
		IdleCheckFrequency: time.Duration(conf.IdleCheckFrequency) * time.Second,
		MaxConnAge:         time.Duration(conf.MaxConnAge) * time.Second,
		PoolTimeout:        time.Duration(conf.PoolTimeout) * time.Second,
	}
	return redis.NewClient(opt)
}

// standalone 标准模式客户端
func standalone(conf config.RedisConfiguration) *redis.Client {
	if conf.Host == "" {
		conf.Host = "proton-redis-proton-redis.resource.svc.cluster.local"
	}
	if conf.Port == "" {
		conf.Port = "6379"
	}
	opt := &redis.Options{
		Addr:               conf.Host + ":" + conf.Port,
		Password:           conf.Password,
		DB:                 conf.DB,
		MaxRetries:         conf.MaxRetries,
		PoolSize:           conf.PoolSize,
		ReadTimeout:        time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout:       time.Duration(conf.WriteTimeout) * time.Second,
		IdleTimeout:        time.Duration(conf.IdleTimeout) * time.Second,
		IdleCheckFrequency: time.Duration(conf.IdleCheckFrequency) * time.Second,
		MaxConnAge:         time.Duration(conf.MaxConnAge) * time.Second,
		PoolTimeout:        time.Duration(conf.PoolTimeout) * time.Second,
	}
	return redis.NewClient(opt)
}

// sentinel 哨兵模式客户端
func sentinel(conf config.RedisConfiguration) *redis.Client {
	if conf.MasterGroupName == "" {
		conf.MasterGroupName = "mymaster"
	}
	if conf.SentinelPwd == "" {
		conf.SentinelPwd = "eisoo.com123"
	}
	if conf.SentinelHost == "" {
		conf.SentinelHost = "proton-redis-proton-redis-sentinel.resource.svc.cluster.local"
	}
	if conf.SentinelPort == "" {
		conf.SentinelPort = "26379"
	}
	opt := redis.FailoverOptions{
		MasterName:         conf.MasterGroupName,
		SentinelAddrs:      []string{fmt.Sprintf("%v:%v", conf.SentinelHost, conf.SentinelPort)},
		SentinelPassword:   conf.SentinelPwd,
		Username:           conf.UserName,
		Password:           conf.SentinelPwd,
		DB:                 conf.DB,
		MaxRetries:         conf.MaxRetries,
		PoolSize:           conf.PoolSize,
		ReadTimeout:        time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout:       time.Duration(conf.WriteTimeout) * time.Second,
		IdleTimeout:        time.Duration(conf.IdleTimeout) * time.Second,
		IdleCheckFrequency: time.Duration(conf.IdleCheckFrequency) * time.Second,
		MaxConnAge:         time.Duration(conf.MaxConnAge) * time.Second,
		PoolTimeout:        time.Duration(conf.PoolTimeout) * time.Second,
	}
	return redis.NewFailoverClient(&opt)
}

// cluster 集群模式客户端
func cluster(conf config.RedisConfiguration) *redis.ClusterClient {
	if conf.ClusterPwd == "" {
		conf.ClusterPwd = "eisoo.com123"
	}
	opt := redis.ClusterOptions{
		Addrs:              conf.ClusterHosts,
		Password:           conf.ClusterPwd,
		MaxRetries:         conf.MaxRetries,
		PoolSize:           conf.PoolSize,
		ReadTimeout:        time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout:       time.Duration(conf.WriteTimeout) * time.Second,
		IdleTimeout:        time.Duration(conf.IdleTimeout) * time.Second,
		IdleCheckFrequency: time.Duration(conf.IdleCheckFrequency) * time.Second,
		MaxConnAge:         time.Duration(conf.MaxConnAge) * time.Second,
		PoolTimeout:        time.Duration(conf.PoolTimeout) * time.Second,
	}
	return redis.NewClusterClient(&opt)
}
