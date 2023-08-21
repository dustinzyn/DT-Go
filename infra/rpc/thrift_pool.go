package rpc

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/rpc/thrift"
)

const (
	CHECKINTERVAL = 120 //清除超时连接间隔

	poolOpen = 1
	poolStop = 2

	DEFAULT_MAX_CONN       = 60
	DEFAULT_CONN_TIMEOUT   = time.Second * 2
	DEFAULT_SOCKET_TIMEOUT = time.Second * 60
	DEFAULT_IDLE_TIMEOUT   = time.Minute * 15
	maxInitConnCount       = 10
	DEFAULT_TIMEOUT        = time.Second * 5
	defaultInterval        = time.Millisecond * 50
)

var (
	ErrOverMax          = fmt.Errorf("ThriftPool connection exceeds the set maximum number of connections")
	ErrInvalidConn      = fmt.Errorf("ThriftPool connection is nil")
	ErrPoolClosed       = fmt.Errorf("ThriftPool has been closed")
	ErrSocketDisconnect = fmt.Errorf("ThriftPool client socket connection has been disconnected")
)

type ThriftDial func(newClientProtocolFunc, clientPtrPtr interface{}, addr string) (*IdleClient, error)

type ThriftClientClose func(*IdleClient) error

// ThriftPool thrift客户端连接池
type ThriftPool struct {
	// 创建client 由使用方注册
	Dial ThriftDial
	// 关闭client 由使用方注册
	Close ThriftClientClose
	// 空闲队列 doubly linked list
	idle list.List
	lock *sync.Mutex
	// 连接数
	count int32
	// 连接池状态
	status int32
	config *ThriftPoolConfig
}

// ThriftPoolConfig .
type ThriftPoolConfig struct {
	// Server addr
	Addr string
	ClientProtocolFunc interface{}
	ClientPtrPtr interface{}
	// 最大连接数
	MaxConn int32
	// 客户端尝试连接到Thrift服务器的超时时间
	ConnTimeout time.Duration
	// 在已建立连接的情况下，客户端发送请求并等待服务器响应的最大时间
	SocketTimeout time.Duration
	// 空闲连接的超时时间，超时会主动释放
	IdleTimeout time.Duration
	// 获取client的超时时间
	Timeout time.Duration
	// 获取client失败的重试间隔
	interval time.Duration
}

// IdleClient thrift客户端
type IdleClient struct {
	// Thrift transport
	Transport thrift.TTransport
	// Thrift client
	RawClient interface{}
}

// 封装的thrift客户端
type idleConn struct {
	// 空闲的客户端
	c *IdleClient
	// 最后放入空闲队列的时间
	t time.Time
}

func NewThriftPool(config *ThriftPoolConfig, dial ThriftDial, closeFunc ThriftClientClose) *ThriftPool {
	// 检查连接池配置
	checkThriftConfig(config)
	thriftPool := &ThriftPool{
		Dial:   dial,
		Close:  closeFunc,
		lock:   &sync.Mutex{},
		config: config,
		status: poolOpen,
		count:  0,
	}
	// 初始化空闲链接
	thriftPool.initConn()
	// 定期清理过期空闲连接
	go thriftPool.ClearConn()
	return thriftPool
}

func checkThriftConfig(config *ThriftPoolConfig) {
	if config.MaxConn == 0 {
		config.MaxConn = DEFAULT_MAX_CONN
	}
	if config.ConnTimeout == 0 {
		config.ConnTimeout = DEFAULT_CONN_TIMEOUT
	}
	if config.SocketTimeout == 0 {
		config.SocketTimeout = DEFAULT_SOCKET_TIMEOUT
	}
	if config.IdleTimeout <= 0 {
		config.IdleTimeout = DEFAULT_IDLE_TIMEOUT
	}
	if config.Timeout <= 0 {
		config.Timeout = DEFAULT_TIMEOUT
	}
	config.interval = defaultInterval
}

// Get 获取空闲客户端
func (p *ThriftPool) Get() (*IdleClient, error) {
	return p.get(time.Now().Add(p.config.Timeout))
}

// expire设定了一个超时时间点，当没有可用连接时，程序会休眠一小段时间后重试
// 如果一直获取不到连接，一旦到达超时时间点，则报ErrOverMax错误
func (p *ThriftPool) get(expire time.Time) (*IdleClient, error) {
	if atomic.LoadInt32(&p.status) == poolStop {
		return nil, ErrPoolClosed
	}

	p.lock.Lock()
	// 无空闲连接 并且连接数超额
	if p.idle.Len() == 0 && atomic.LoadInt32(&p.count) >= p.config.MaxConn {
		p.lock.Unlock()
		for {
			time.Sleep(p.config.interval)
			// 超时退出
			if time.Now().After(expire) {
				return nil, ErrOverMax
			}
			p.lock.Lock()
			if p.idle.Len() == 0 && atomic.LoadInt32(&p.count) >= p.config.MaxConn {
				p.lock.Unlock()
			} else {
				// 获取到可用连接
				break
			}
		}
	}

	if p.idle.Len() == 0 {
		// 首次创建连接，先加1
		atomic.AddInt32(&p.count, 1)
		p.lock.Unlock()
		// 创建连接
		client, err := p.Dial(p.config.ClientProtocolFunc, p.config.ClientPtrPtr, p.config.Addr)
		if err != nil {
			atomic.AddInt32(&p.count, -1)
			return nil, err
		}
		if !client.Check() {
			atomic.AddInt32(&p.count, -1)
			return nil, ErrSocketDisconnect
		}

		return client, nil
	}

	// 从队列头部获取空闲连接
	element := p.idle.Front()
	idlec := element.Value.(*idleConn)
	p.idle.Remove(element)
	p.lock.Unlock()

	if !idlec.c.Check() {
		atomic.AddInt32(&p.count, -1)
		return nil, ErrSocketDisconnect
	}
	return idlec.c, nil
}

// 归还Thrift客户端至连接池
func (p *ThriftPool) Put(client *IdleClient) error {
	if client == nil {
		return nil
	}

	if atomic.LoadInt32(&p.status) == poolStop {
		err := p.Close(client)
		client = nil
		return err
	}

	if atomic.LoadInt32(&p.count) > p.config.MaxConn || !client.Check() {
		atomic.AddInt32(&p.count, -1)
		err := p.Close(client)
		client = nil
		return err
	}

	p.lock.Lock()
	p.idle.PushFront(&idleConn{
		c: client,
		t: time.Now(),
	})
	p.lock.Unlock()

	return nil
}

// 关闭异常连接 创建新的连接
func (p *ThriftPool) Reconnect(client *IdleClient) (newClient *IdleClient, err error) {
	if client != nil {
		p.Close(client)
	}
	// client = nil

	newClient, err = p.Dial(p.config.ClientProtocolFunc, p.config.ClientPtrPtr, p.config.Addr)
	if err != nil {
		atomic.AddInt32(&p.count, -1)
		return
	}
	if !newClient.Check() {
		atomic.AddInt32(&p.count, -1)
		return nil, ErrSocketDisconnect
	}
	return
}

// 关闭连接
func (p *ThriftPool) CloseConn(client *IdleClient) {
	if client != nil {
		p.Close(client)
	}
	atomic.AddInt32(&p.count, -1)
}

func (p *ThriftPool) CheckTimeout() {
	p.lock.Lock()
	for p.idle.Len() != 0 {
		ele := p.idle.Back()
		if ele == nil {
			break
		}
		v := ele.Value.(*idleConn)
		if v.t.Add(p.config.IdleTimeout).After(time.Now()) {
			break
		}

		// 超时了直接移除
		p.idle.Remove(ele)
		p.lock.Unlock()

		p.Close(v.c)
		atomic.AddInt32(&p.count, -1)

		p.lock.Lock()
	}
	p.lock.Unlock()
}

// 检测连接是否有效
func (c *IdleClient) Check() bool {
	if c.Transport == nil || c.RawClient == nil {
		return false
	}
	return c.Transport.IsOpen()
}

func (p *ThriftPool) GetIdleCount() int32 {
	if p != nil {
		return int32(p.idle.Len())
	}
	return 0
}

func (p *ThriftPool) GetConnCount() int32 {
	if p != nil {
		return atomic.LoadInt32(&p.count)
	}
	return 0
}

func (p *ThriftPool) ClearConn() {
	// gap := CHECKINTERVAL * time.Second
	// if gap < p.config.IdleTimeout {
	// 	gap = p.config.IdleTimeout
	// }
	for {
		p.CheckTimeout()
		time.Sleep(CHECKINTERVAL * time.Second)
	}
}

// 释放所有连接
func (p *ThriftPool) Release() {
	atomic.StoreInt32(&p.status, poolStop)
	atomic.StoreInt32(&p.count, 0)

	p.lock.Lock()
	idle := p.idle
	p.idle.Init()
	p.lock.Unlock()

	for item := idle.Front(); item != nil; item = item.Next() {
		p.Close(item.Value.(*idleConn).c)
	}
}

func (p *ThriftPool) Recover() {
	atomic.StoreInt32(&p.status, poolOpen)
}

// 创建连接池，并初始化一定数量的连接
func (p *ThriftPool) initConn() {
	initCount := p.config.MaxConn
	if initCount > maxInitConnCount {
		initCount = maxInitConnCount
	}
	wg := &sync.WaitGroup{}
	wg.Add(int(initCount))
	for i := int32(0); i < initCount; i++ {
		go p.createIdleConn(wg)
	}
	wg.Wait()
}

func (p *ThriftPool) createIdleConn(wg *sync.WaitGroup) {
	c, _ := p.Get()
	p.Put(c)
	wg.Done()
}
