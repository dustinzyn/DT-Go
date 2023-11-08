// 消息代理客户端组件
package mqclient

import (
	"strconv"

	dhive "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	msqclient "devops.aishu.cn/AISHUDevOps/ONE-Architecture/_git/proton-mq-go"
)

//go:generate mockgen -package mock_infra -source protonMqClient.go -destination ./mock/protonmq_mock.go

func init() {
	dhive.Prepare(func(initiator dhive.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *ProtonMQClientImpl {
			return &ProtonMQClientImpl{}
		})
	})
}

type ProtonMQClient interface {
	// Pub send a message to the specified topic of msq
	Pub(topic string, msg []byte) error
	// Sub start consumers to subscribe and process message from specified topic/channel from the msg, the call would run
	// forever until the program is terminated
	Sub(topic string, channel string, handler func([]byte) error, pollIntervalMilliseconds int64, maxInFlight int) error

	Close()
}

type ProtonMQClientImpl struct {
	dhive.Infra
	producerHost  string
	producerPort  int
	consumerHost  string
	consumerPort  int
	connectorType string
	mqClient      msqclient.ProtonMQClient
}

func (mq *ProtonMQClientImpl) newClient() {
	var err error
	cg := dhive.NewConfiguration()
	mq.producerHost = cg.MQ.ProducerHost
	producerPort := cg.MQ.ProducerPort
	pport, _ := strconv.Atoi(producerPort)
	mq.producerPort = pport

	mq.consumerHost = cg.MQ.ConsumerHost
	consumerPort := cg.MQ.ConsumerPort
	cport, _ := strconv.Atoi(consumerPort)
	mq.consumerPort = cport

	mq.connectorType = cg.MQ.ConnectType
	mq.mqClient, err = msqclient.NewProtonMQClient(
		mq.producerHost, mq.producerPort,
		mq.consumerHost, mq.consumerPort,
		mq.connectorType,
	)
	if err != nil {
		panic(err)
	}
}

func (mq *ProtonMQClientImpl) BeginRequest(worker dhive.Worker) {
	mq.newClient()
	mq.Infra.BeginRequest(worker)
}

func (mq *ProtonMQClientImpl) Begin() {
	mq.newClient()
}

func (mq *ProtonMQClientImpl) Pub(topic string, msg []byte) error {
	return mq.mqClient.Pub(topic, msg)
}

func (mq *ProtonMQClientImpl) Sub(topic string, channel string, handler func([]byte) error, pollIntervalMilliseconds int64, maxInFlight int) error {
	return mq.mqClient.Sub(topic, channel, handler, pollIntervalMilliseconds, maxInFlight)
}

func (mq *ProtonMQClientImpl) Close() {
	mq.mqClient.Close()
}
