// 消息代理客户端组件
package mqclient

import (
	"os"
	"strconv"

	hive "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	msqclient "devops.aishu.cn/AISHUDevOps/ICT/_git/go-msq"
)

func init() {
	hive.Prepare(func(initiator hive.Initiator) {
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
	hive.Infra
	producerHost  string
	producerPort  int
	consumerHost  string
	consumerPort  int
	connectorType string
	mqClient      msqclient.ProtonMSQClient
}

func (mq *ProtonMQClientImpl) newClient() {
	mq.producerHost = os.Getenv("MQ_PRODUCER_HOST")
	producerPort := os.Getenv("MQ_PRODUCER_PORT")
	pport, _ := strconv.Atoi(producerPort)
	mq.producerPort = pport

	mq.consumerHost = os.Getenv("MQ_CONSUMER_HOST")
	consumerPort := os.Getenv("MQ_CONSUMER_PORT")
	cport, _ := strconv.Atoi(consumerPort)
	mq.consumerPort = cport

	mq.connectorType = os.Getenv("MQ_CONNECTOR_TYPE")
	mq.mqClient = msqclient.NewProtonMSQClient(
		mq.producerHost, mq.producerPort,
		mq.consumerHost, mq.consumerPort,
		mq.connectorType,
	)
}

func (mq *ProtonMQClientImpl) BeginRequest(worker hive.Worker) {
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
