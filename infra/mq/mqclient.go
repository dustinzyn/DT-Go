package mqclient

import (
	"os"
	"strconv"
	"sync"

	msqclient "devops.aishu.cn/AISHUDevOps/ICT/_git/go-msq"
)

var (
	producerHost  string = os.Getenv("MQ_PRODUCER_HOST")
	producerPort  string = os.Getenv("MQ_PRODUCER_PORT")
	consumerHost  string = os.Getenv("MQ_CONSUMER_HOST")
	consumerPort  string = os.Getenv("MQ_CONSUMER_PORT")
	connectorType string = os.Getenv("MQ_CONNECTOR_TYPE")
	mqClient      msqclient.ProtonMSQClient
	mqOnce        sync.Once
)

func NewMQClient() msqclient.ProtonMSQClient {
	mqOnce.Do(func() {
		pport, _ := strconv.Atoi(producerPort)
		cport, _ := strconv.Atoi(consumerPort)
		mqClient = msqclient.NewProtonMSQClient(producerHost, pport, consumerHost, cport, connectorType)
	})
	return mqClient
}
