package store

/**
缓存组件
实现了一级缓存 二级缓存 集成redis
基于sync/singlefight实现缓存防击穿 Service->Cache->Singleflight->DB

Created by Dustin.zhu on 2022/11/1.
*/

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	redis "github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
)

// EntityCache .
type EntityCache interface {
	// 获取实体
	GetEntity(hive.Entity) error
	// 删除实体缓存
	Delete(result hive.Entity, async ...bool) error
	// 设置数据源
	SetSource(func(hive.Entity) error) EntityCache
	// 设置前缀
	SetPrefix(string) EntityCache
	// 设置缓存时间 默认5分钟
	SetExpiration(time.Duration) EntityCache
	// 设置异步反写缓存 默认关闭 缓存未命中读取数据源后的异步反写缓存
	SetAsyncWrite(bool) EntityCache
	// 设置防击穿 默认开启
	SetSingleFlight(bool) EntityCache
	// 关闭二级缓存 关闭后只有一级缓存生效
	CloseSecondCache() EntityCache
}

func init() {
	hive.Prepare(func(initiator hive.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *EntityCacheImpl {
			return &EntityCacheImpl{}
		})
	})
}

var sfGroup singleflight.Group

// EntityCacheImpl .
type EntityCacheImpl struct {
	hive.Infra                                  // Infra
	asyncWrite   bool                           // 异步写缓存
	prefix       string                         // 缓存前缀
	expiration   time.Duration                  // 缓存有效期
	call         func(result hive.Entity) error // 未命中缓冲的回调函数
	singleFlight bool                           // 缓存防击穿
	client       redis.Cmdable                  // redis client
}

// BeginRequest .
func (cache *EntityCacheImpl) BeginRequest(worker hive.Worker) {
	cache.expiration = 5 * time.Minute
	cache.singleFlight = true
	cache.asyncWrite = false
	cache.Infra.BeginRequest(worker)
	cache.client = cache.Redis()
}

// GetEntity .
func (cache *EntityCacheImpl) GetEntity(result hive.Entity) error {
	value := reflect.ValueOf(result)
	name := cache.getName(value.Type()) + ":" + result.Identity()
	// 读取一级缓存
	ok, err := cache.getStore(name, result)
	if err != nil || ok {
		return err
	}

	// 读取二级缓存
	entityBytes, err := cache.getRedis(name)
	if err != nil || err != redis.Nil {
		return err
	}
	if err != redis.Nil {
		err = json.Unmarshal(entityBytes, result)
		if err != nil {
			return err
		}
		// 刷新一级缓存
		cache.setStore(name, entityBytes)
	}

	// 未命中缓存 读取DB
	entityBytes, err = cache.getCall(name, result)
	if err != nil {
		return err
	}

	// 反写缓存
	cache.setStore(name, entityBytes)
	if cache.client == nil {
		return nil
	}
	if !cache.asyncWrite {
		return cache.client.Set(cache.Worker().Context(), name, entityBytes, cache.expiration).Err()
	}
	go func() {
		var err error
		defer func() {
			if perr := recover(); perr != nil {
				err = fmt.Errorf(fmt.Sprint(perr))
			}
			if err != nil {
				hive.Logger().Errorf("Failed to set entity cache, name:%s, err:%v", name, err)
			}
		}()
		err = cache.client.Set(cache.Worker().Context(), name, entityBytes, cache.expiration).Err()
	}()

	return nil
}

// Delete delete entity cache.
func (cache *EntityCacheImpl) Delete(result hive.Entity, async ...bool) error {
	name := cache.getName(reflect.ValueOf(result).Type()) + ":" + result.Identity()
	if !cache.Worker().IsDeferRecycle() {
		cache.Worker().Store().Remove(name)
	}
	client := cache.client
	if client == nil {
		return nil
	}
	if len(async) == 0 {
		return client.Del(cache.Worker().Context(), name).Err()
	}
	go func() {
		var err error
		defer func() {
			if perr := recover(); perr != nil {
				err = fmt.Errorf(fmt.Sprint(perr))
			}
			if err != nil {
				hive.Logger().Errorf("Failed to delete entity cache, name:%s, err:%v", name, err)
			}
		}()
		err = client.Del(cache.Worker().Context(), name).Err()
	}()
	return nil
}

// SetSource 设置数据源
func (cache *EntityCacheImpl) SetSource(call func(result hive.Entity) error) EntityCache {
	cache.call = call
	return cache
}

// SetAsyncWrite 设置异步写入缓存 缓存未命中时 读取数据源后异步写入缓存
func (cache *EntityCacheImpl) SetAsyncWrite(open bool) EntityCache {
	cache.asyncWrite = open
	return cache
}

// SetExpiration 设置缓存时间 默认为5分钟
func (cache *EntityCacheImpl) SetExpiration(expiration time.Duration) EntityCache {
	cache.expiration = expiration
	return cache
}

// SetPrefix 设置缓存前缀
func (cache *EntityCacheImpl) SetPrefix(prefix string) EntityCache {
	cache.prefix = prefix
	return cache
}

// SetSingleFlight 设置是否开启防击穿 默认开启
func (cache *EntityCacheImpl) SetSingleFlight(open bool) EntityCache {
	cache.singleFlight = open
	return cache
}

// CloseSecondCache 关闭二级缓存
func (cache *EntityCacheImpl) CloseSecondCache() EntityCache {
	cache.client = nil
	return cache
}

func (cache *EntityCacheImpl) getName(entityType reflect.Type) string {
	for entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	if cache.prefix != "" {
		return cache.prefix + ":" + entityType.Name()
	}
	return entityType.Name()
}

func (cache *EntityCacheImpl) getStore(name string, result hive.Entity) (bool, error) {
	if cache.Worker().IsDeferRecycle() {
		return false, nil
	}
	entityStore := cache.Worker().Store().Get(name)
	if entityStore == nil {
		return false, nil
	}
	if err := json.Unmarshal(entityStore.([]byte), result); err != nil {
		return false, err
	}
	return true, nil
}

func (cache *EntityCacheImpl) getRedis(name string) ([]byte, error) {
	if cache.client == nil {
		return nil, redis.Nil
	}
	client := cache.client
	if cache.singleFlight {
		entityData, err, _ := sfGroup.Do("cache:"+name, func() (interface{}, error) {
			return client.Get(cache.Worker().Context(), name).Bytes()
		})
		if err != nil {
			return nil, err
		}
		return entityData.([]byte), err
	}
	return client.Get(cache.Worker().Context(), name).Bytes()
}

func (cache *EntityCacheImpl) setStore(name string, store []byte) {
	if cache.Worker().IsDeferRecycle() {
		return
	}
	cache.Worker().Store().Set(name, store)
}

func (cache *EntityCacheImpl) getCall(name string, result hive.Entity) ([]byte, error) {
	if cache.call == nil {
		return nil, errors.New("Undefined source")
	}

	if cache.singleFlight {
		var err error
		sonCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		res := sfGroup.DoChan("call:"+name, func() (interface{}, error) {
			e := cache.call(result)
			if e != nil {
				return nil, e
			}
			return json.Marshal(result)
		})
		select {
		case r := <-res:
			if r.Err != nil {
				return nil, r.Err
			}
			entityData := r.Val
			if r.Shared {
				entityDataByte, _ := entityData.([]byte)
				err = json.Unmarshal(entityDataByte, result)
				return entityDataByte, err
			} else {
				return entityData.([]byte), err
			}
		case <-sonCtx.Done():
			sfGroup.Forget("call:" + name)
			return nil, errors.New("getCall timeout")
		}
	}
	err := cache.call(result)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
