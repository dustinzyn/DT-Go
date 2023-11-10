package dlm

import (
	"context"
	"fmt"
	"sync"
	"time"

	dt "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/DT-Go"
	redis "github.com/go-redis/redis/v8"
)

/**
分布式互斥锁组件
	1.原子性：利用 Lua 脚本实现原子性语义；
	2.锁自动过期：避免因为宕机导致的死锁问题；
	3.锁的自动续期：利用 Go 协程实现锁资源的自动续期(看门狗机制)，避免出现业务时间>锁超时时间导致并发安全问题
	4.TryLock机制：仅尝试一次锁的获取，如果失败，那么不会阻塞，直接返回
	5.自旋锁：提供自旋锁 API 来实现分布式锁的自旋获取
不支持如下特性：
	1.重入性：分布式锁不可重入，Go 语言并没有优雅的方式来实现 Java 中的 ThreadLocal 机制
	2.非公平性：分布式锁存在非公平问题，在极端情况下会导致饥饿问题

Created by Dustin.zhu on 2023/05/10.
*/

//go:generate mockgen -package mock_infra -source distributed_lock.go -destination ./mock/dlm_mock.go

func init() {
	dt.Prepare(func(initiator dt.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *DLMImpl {
			return &DLMImpl{}
		})
		mutex = &sync.Mutex{}
	})
}

const (
	DefaultExpiration      = 30 // second
	DefaultSpinInterval    = 100
	DefaultTryLockInterval = 20 // millisecond
)

var mutex *sync.Mutex

// DLM .
type DLM interface {
	// SetExpiration Set a timeout period with a default duration of 30 second.
	SetExpiration(time.Duration) DLM
	// SetContext The timeout for obtaining locks can be controlled through context
	SetContext(ctx context.Context) DLM
	// Lock blocked until get lock
	Lock(key string) error
	// TryLock try get lock only once, if get the lock return true, else return false
	TryLock(key string) (bool, error)
	// SpinLock is a synchronization mechanism that allows multiple threads to
	// access shared resources without blocking, by continuously executing
	// a tight loop of instructions ("spinning") while waiting for the lock to become available.
	SpinLock(key string, times int) error
	// Unlock release the lock
	Unlock(key string) error
}

// DLMImpl .
type DLMImpl struct {
	dt.Infra   // Infra
	value      string
	client     redis.Cmdable
	expiration time.Duration
	cancelFunc context.CancelFunc
	ctx        context.Context
}

// BeginRequest .
func (dlm *DLMImpl) BeginRequest(worker dt.Worker) {
	dlm.expiration = DefaultExpiration * time.Second
	dlm.client = dlm.Redis()
	dlm.ctx = context.Background()
	dlm.value = time.Now().GoString()
	dlm.Infra.BeginRequest(worker)
}

// SetExpiration .
func (dlm *DLMImpl) SetExpiration(ex time.Duration) DLM {
	dlm.expiration = ex
	return dlm
}

// SetContext .
func (dlm *DLMImpl) SetContext(ctx context.Context) DLM {
	dlm.ctx = ctx
	return dlm
}

// Lock blocked until get lock.
func (dlm *DLMImpl) Lock(key string) (err error) {
	// 尝试获取锁
	ok, err := dlm.TryLock(key)
	if err != nil {
		return
	}
	if ok {
		return
	}
	ticker := time.NewTicker(DefaultTryLockInterval * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			// 重新尝试获取锁
			ok, err = dlm.TryLock(key)
			if err != nil {
				return
			}
			if ok {
				return
			}
		case <-dlm.ctx.Done():
			return fmt.Errorf("get lock timeout.")
		}
	}
}

// TryLock try get lock only once, if get the lock return true, else return false.
func (dlm *DLMImpl) TryLock(key string) (ok bool, err error) {
	mutex.Lock()
	defer mutex.Unlock()
	ok, err = dlm.client.SetNX(dlm.ctx, key, dlm.value, dlm.expiration).Result()
	if err != nil {
		return
	}
	// 使用context控制watchdog协程的执行和取消
	cancelCtx, cancelFunc := context.WithCancel(dlm.ctx)
	dlm.cancelFunc = cancelFunc
	go dlm.watchdog(cancelCtx, key)
	return
}

// watchdog will renew the expiration of lock, and can be canceled when call Unlock
func (dlm *DLMImpl) watchdog(ctx context.Context, key string) {
	ticker := time.NewTicker(dlm.expiration / 3)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			dlm.client.Expire(ctx, key, dlm.expiration).Result()
		}
	}
}

// SpinLock .
func (dlm *DLMImpl) SpinLock(key string, times int) error {
	for i := 0; i < times; i++ {
		ok, err := dlm.TryLock(key)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		time.Sleep(time.Microsecond * DefaultSpinInterval)
	}
	return fmt.Errorf("max spin times reached.")
}

// Unlock .
func (dlm *DLMImpl) Unlock(key string) error {
	script := redis.NewScript(fmt.Sprintf(
		`if redis.call("get", KEYS[1]) == "%s" then return redis.call("del", KEYS[1]) else return 0 end`,
		dlm.value))
	runCmd := script.Run(context.Background(), dlm.client, []string{key})
	res, err := runCmd.Result()
	if err != nil {
		return err
	}
	if val, ok := res.(int64); ok {
		if val == 1 {
			dlm.cancelFunc() // 结束renew协程
			return nil
		}
	}
	err = fmt.Errorf("unlock script fail: %v", key)
	return err
}
