// 消息代理客户端组件
package mqclient

import (
	"strconv"

	dt "DT-Go"
)

//go:generate mockgen -package mock_infra -source protonMqClient.go -destination ./mock/protonmq_mock.go

type NewClienFunc func(pubServer string, pubPort int, subServer string, subPort int) MQClient

// mq client factory
var mqcFactory map[string]NewClienFunc

type MsgHandler func([]byte) error

func init() {
	mqcFactory = make(map[string]NewClienFunc, 5)
	mqcFactory["nsq"] = NewNSQClient
	mqcFactory["kafka"] = NewKafkaClient
	dt.Prepare(func(initiator dt.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *MQClientImpl {
			return &MQClientImpl{}
		})
	})
}

type MQClient interface {
	// Pub send a message to the specified topic of msq
	Pub(topic string, msg []byte) error
	// Sub start consumers to subscribe and process message from specified topic/channel from the msg, the call would run
	// forever until the program is terminated
	Sub(topic string, channel string, handler func([]byte) error, pollIntervalMilliseconds int64, maxInFlight int) error

	Close()
}

type MQClientImpl struct {
	dt.Infra
	producerHost  string
	producerPort  int
	consumerHost  string
	consumerPort  int
	connectorType string
	mqClient      MQClient
}

func (mq *MQClientImpl) newClient() {
	cg := dt.NewConfiguration()
	mq.producerHost = cg.MQ.ProducerHost
	producerPort := cg.MQ.ProducerPort
	pport, _ := strconv.Atoi(producerPort)
	mq.producerPort = pport

	mq.consumerHost = cg.MQ.ConsumerHost
	consumerPort := cg.MQ.ConsumerPort
	cport, _ := strconv.Atoi(consumerPort)
	mq.consumerPort = cport

	mq.connectorType = cg.MQ.ConnectType

	// create mq client for specified mq type
	if fn, ok := mqcFactory[mq.connectorType]; ok {
		mq.mqClient = fn(mq.producerHost, mq.producerPort, mq.consumerHost, mq.consumerPort)
		return
	} else {
		panic("not supported mq type: " + mq.connectorType)
	}
}

func (mq *MQClientImpl) BeginRequest(worker dt.Worker) {
	mq.newClient()
	mq.Infra.BeginRequest(worker)
}

func (mq *MQClientImpl) Begin() {
	mq.newClient()
}

func (mq *MQClientImpl) Pub(topic string, msg []byte) error {
	return mq.mqClient.Pub(topic, msg)
}

func (mq *MQClientImpl) Sub(topic string, channel string, handler func([]byte) error, pollIntervalMilliseconds int64, maxInFlight int) error {
	return mq.mqClient.Sub(topic, channel, handler, pollIntervalMilliseconds, maxInFlight)
}

func (mq *MQClientImpl) Close() {
	mq.mqClient.Close()
}
