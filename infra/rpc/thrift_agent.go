package rpc

import (
	"fmt"
	"net"
	"reflect"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/DT-Go/infra/rpc/thrift"
)

type ThriftPoolAgent struct {
	pool *ThriftPool
}

func NewThriftPoolAgent(config *ThriftPoolConfig) *ThriftPoolAgent {
	pool := NewThriftPool(config, thriftDial, thriftClientClose)
	poolAgent := &ThriftPoolAgent{}
	poolAgent.init(pool)
	return poolAgent
}

func (a *ThriftPoolAgent) init(pool *ThriftPool) {
	a.pool = pool
}

func thriftDial(newClientProtocolFunc, clientPtrPtr interface{}, addr string) (*IdleClient, error) {
	socket, err := thrift.NewTSocket(addr)
	if err != nil {
		return nil, err
	}

	transport := thrift.NewTBufferedTransport(socket, 8192)
	if err := transport.Open(); err != nil {
		return nil, err
	}

	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	argsV := make([]reflect.Value, 2)
	argsV[0] = reflect.ValueOf(transport)
	argsV[1] = reflect.ValueOf(protocolFactory)
	clientV := reflect.ValueOf(newClientProtocolFunc).Call(argsV)
	reflect.ValueOf(clientPtrPtr).Elem().Set(clientV[0])
	return &IdleClient{
		Transport: transport,
		RawClient: clientPtrPtr,
	}, nil
}

func thriftClientClose(c *IdleClient) error {
	if c == nil {
		return nil
	}
	return c.Transport.Close()
}

func (a *ThriftPoolAgent) Do(do func(rawClient interface{}) error) error {
	var (
		client *IdleClient
		err    error
	)
	defer func() {
		if client == nil {
			return
		}
		if _, ok := err.(net.Error); ok {
			a.closeClient(client)
		} else if _, ok = err.(thrift.TTransportException); ok {
			a.closeClient(client)
		} else {
			if rErr := a.releaseClient(client); rErr != nil {
				fmt.Printf("[error] releaseClient: %v", rErr)
			}
		}
	}()

	client, err = a.getClient()
	if err != nil {
		return err
	}
	if err = do(client.RawClient); err != nil {
		if _, ok := err.(net.Error); ok {
			fmt.Printf("[error] retry tcp, %T, %s", err, err.Error())
			client, err = a.reconnect(client)
			if err != nil {
				return err
			}
			return do(client.RawClient)
		}
		if _, ok := err.(thrift.TTransportException); ok {
			fmt.Printf("[error] retry tcp: %T, %s", err, err.Error())
			client, err = a.reconnect(client)
			if err != nil {
				return err
			}
			return do(client.RawClient)
		}
		return err
	}
	return nil
}

// 获取连接
func (a *ThriftPoolAgent) getClient() (*IdleClient, error) {
	return a.pool.Get()
}

// 释放连接
func (a *ThriftPoolAgent) releaseClient(client *IdleClient) error {
	return a.pool.Put(client)
}

// 关闭有问题的连接，并重新创建一个新的连接
func (a *ThriftPoolAgent) reconnect(client *IdleClient) (newClient *IdleClient, err error) {
	return a.pool.Reconnect(client)
}

// 关闭连接
func (a *ThriftPoolAgent) closeClient(client *IdleClient) {
	a.pool.CloseConn(client)
}

// 释放连接池
func (a *ThriftPoolAgent) Release() {
	a.pool.Release()
}

func (a *ThriftPoolAgent) GetIdleCount() int32 {
	return a.pool.GetIdleCount()
}

func (a *ThriftPoolAgent) GetConnCount() int32 {
	return a.pool.GetConnCount()
}
