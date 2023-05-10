package store

/*
	分布式缓存组件
	实现了一级缓存 二级缓存 集成redis
	一级缓存的生命周期为单个请求，仅用于函数间调用
	基于sync/singlefight实现缓存防击穿 Service->Cache->Singleflight->DB

	采用 Cache-Aside Pattern （旁路缓存模式）机制实现
	读：读的时候，先读缓存，缓存命中，直接返回数据。缓存没有命中，就去读数据库，然后用数据库的数据更新缓存，再返回数据
	写：更新数据的时候，先更新数据库，再删除缓存

	Created by Dustin.zhu on 2023/04/24.
*/

import (
	"context"
	"errors"
	"fmt"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	redis "github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"
)

func init() {
	hive.Prepare(func(initiator hive.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *CacheImpl {
			return &CacheImpl{}
		})
	})
}

// Cache .
type Cache interface {
	// 获取缓存
	Get(key string, expiration ...time.Duration) (cacheBytes []byte, err error)
	// 删除实体缓存
	Delete(key string, async ...bool) error
	// 设置数据源
	SetSource(func() (cacheBytes []byte, err error)) Cache
	// 设置前缀
	SetPrefix(string) Cache
	// 设置缓存时间 默认5分钟
	SetExpiration(time.Duration) Cache
	// 设置异步反写缓存 默认关闭 缓存未命中读取数据源后的异步反写缓存
	SetAsyncWrite(bool) Cache
	// 设置防击穿 默认开启
	SetSingleFlight(bool) Cache
	// 关闭二级缓存 关闭后只有一级缓存生效
	CloseSecondCache() Cache
}

var sfGroup singleflight.Group

// CacheImpl .
type CacheImpl struct {
	hive.Infra                                         // Infra
	asyncWrite   bool                                  // 异步写缓存
	prefix       string                                // 缓存前缀
	expiration   time.Duration                         // 缓存有效期
	call         func() (cacheBytes []byte, err error) // 未命中缓冲的回调函数
	singleFlight bool                                  // 缓存防击穿
	client       redis.Cmdable                         // redis client
}

// BeginRequest .
func (cache *CacheImpl) BeginRequest(worker hive.Worker) {
	cache.expiration = 5 * time.Minute
	cache.singleFlight = true
	cache.asyncWrite = false
	cache.client = cache.Redis()
	cache.Infra.BeginRequest(worker)
}

// Get .
func (cache *CacheImpl) Get(key string, expiration ...time.Duration) (cacheBytes []byte, err error) {
	// 读取一级缓存
	cacheBytes, err = cache.getStore(key)
	if err != nil {
		return
	}

	if cacheBytes != nil {
		cache.Worker().Logger().Infof("fetched memstore, key=%v", key)
		return
	}

	// 读取二级缓存
	cacheBytes, err = cache.getRedis(key)
	if err != redis.Nil && err != nil {
		return
	}
	if err != redis.Nil {
		cache.Worker().Logger().Infof("fetched redis cache, key=%v", key)
		// 刷新一级缓存
		cache.setStore(key, cacheBytes)
		return
	}

	// 未命中缓存 读取DB
	cacheBytes, err = cache.getCall(key)
	if err != nil {
		return
	}

	// 反写缓存
	cache.setStore(key, cacheBytes)
	if cache.client == nil {
		return
	}
	var expire time.Duration
	if len(expiration) != 0 {
		expire = expiration[0]
	} else {
		expire = cache.expiration
	}
	if !cache.asyncWrite {
		err = cache.client.Set(cache.Worker().Context(), key, cacheBytes, expire).Err()
		return
	}
	go func() {
		var err error
		defer func() {
			if perr := recover(); perr != nil {
				err = fmt.Errorf(fmt.Sprint(perr))
			}
			if err != nil {
				hive.Logger().Errorf("Failed to set cache, key:%s, err:%v", key, err)
			}
		}()
		err = cache.client.Set(cache.Worker().Context(), key, cacheBytes, expire).Err()
	}()

	return
}

// Delete delete entity cache.
func (cache *CacheImpl) Delete(key string, async ...bool) error {
	if !cache.Worker().IsDeferRecycle() {
		cache.Worker().Store().Remove(key)
	}
	client := cache.client
	if client == nil {
		return nil
	}
	if len(async) == 0 {
		return client.Del(cache.Worker().Context(), key).Err()
	}
	go func() {
		var err error
		defer func() {
			if perr := recover(); perr != nil {
				err = fmt.Errorf(fmt.Sprint(perr))
			}
			if err != nil {
				hive.Logger().Errorf("Failed to delete cache, key:%s, err:%v", key, err)
			}
		}()
		err = client.Del(cache.Worker().Context(), key).Err()
	}()
	return nil
}

// SetSource 设置数据源
func (cache *CacheImpl) SetSource(call func() ([]byte, error)) Cache {
	cache.call = call
	return cache
}

// SetAsyncWrite 设置异步写入缓存 缓存未命中时 读取数据源后异步写入缓存
func (cache *CacheImpl) SetAsyncWrite(open bool) Cache {
	cache.asyncWrite = open
	return cache
}

// SetExpiration 设置缓存时间 默认为5分钟
func (cache *CacheImpl) SetExpiration(expiration time.Duration) Cache {
	cache.expiration = expiration
	return cache
}

// SetPrefix 设置缓存前缀
func (cache *CacheImpl) SetPrefix(prefix string) Cache {
	cache.prefix = prefix
	return cache
}

// SetSingleFlight 设置是否开启防击穿 默认开启
func (cache *CacheImpl) SetSingleFlight(open bool) Cache {
	cache.singleFlight = open
	return cache
}

// CloseSecondCache 关闭二级缓存
func (cache *CacheImpl) CloseSecondCache() Cache {
	cache.client = nil
	return cache
}

func (cache *CacheImpl) getStore(key string) ([]byte, error) {
	if cache.Worker().IsDeferRecycle() {
		return nil, nil
	}
	valueBytes := cache.Worker().Store().Get(key)
	if valueBytes == nil {
		return nil, nil
	}
	_, ok := valueBytes.([]byte)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Invalid cache. key=%v, data=%v", key, valueBytes))
	}
	return valueBytes.([]byte), nil
}

func (cache *CacheImpl) getRedis(key string) ([]byte, error) {
	if cache.client == nil {
		return nil, redis.Nil
	}
	client := cache.client
	ctx := cache.Worker().Context()
	if cache.singleFlight {
		cacheData, err, _ := sfGroup.Do("cache:"+key, func() (interface{}, error) {
			return client.Get(ctx, key).Bytes()
		})
		if err != nil {
			return nil, err
		}
		if cacheData == nil {
			return nil, nil
		}
		_, ok := cacheData.([]byte)
		if !ok {
			return nil, errors.New(fmt.Sprintf("Invalid cache. key=%v, data=%v", key, cacheData))
		}
		return cacheData.([]byte), err
	}
	return client.Get(ctx, key).Bytes()
}

func (cache *CacheImpl) setStore(key string, store []byte) {
	if cache.Worker().IsDeferRecycle() {
		return
	}
	cache.Worker().Store().Set(key, store)
}

func (cache *CacheImpl) getCall(key string) (cacheBytes []byte, err error) {
	if cache.call == nil {
		return nil, errors.New("Undefined source")
	}

	if cache.singleFlight {
		var err error
		sonCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		res := sfGroup.DoChan("call:"+key, func() (interface{}, error) {
			resultBytes, e := cache.call()
			if e != nil {
				return nil, e
			}
			return resultBytes, e
		})
		select {
		case r := <-res:
			if r.Err != nil {
				return nil, r.Err
			}
			cacheData := r.Val
			if cacheData == nil {
				return nil, nil
			}
			return cacheData.([]byte), err
		case <-sonCtx.Done():
			sfGroup.Forget("call:" + key)
			return nil, errors.New("getCall timeout")
		}
	}
	cacheBytes, err = cache.call()
	if err != nil {
		return nil, err
	}
	if cacheBytes == nil {
		return nil, nil
	}
	return
}
