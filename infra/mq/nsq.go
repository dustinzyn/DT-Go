package mqclient

import (
	dt "DT-Go"
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
)


type NSQClient struct {
	dt.Infra
	producerHost string
	producerPort int
	consumerHost string
	consumerPort int
}

func NewNSQClient(pubServer string, pubPort int, subServer string, subPort int) MQClient {
	return &NSQClient{
		producerHost: pubServer,
		producerPort: pubPort,
		consumerHost: subServer,
		consumerPort: subPort,
	}
}

func (nc *NSQClient) pubServer() string{
	return fmt.Sprintf("%s:%d", nc.producerHost, nc.producerPort)
}

func (nc *NSQClient) subServer() string{
	return fmt.Sprintf("%s:%d", nc.consumerHost, nc.consumerPort)
}

// Pub send a message to the specified topic of msq
func (nc *NSQClient) Pub(topic string, msg []byte) error {
	config := nsq.NewConfig()

	producer, err := nsq.NewProducer(nc.pubServer(), config)
	if err != nil {
		nc.Worker().Logger().Errorf("create nsq producer failed, err: %v", err)
		return err
	}
	err = producer.Publish(topic, msg)
	if err != nil {
		nc.Worker().Logger().Errorf("publish message failed, err: %v", err)
		return err
	}
	return nil
}

// Sub start consumers to subscribe and process message from specified topic/channel from the msg, the call would run
// forever until the program is terminated
func (nc *NSQClient) Sub(topic string, channel string, handler func([]byte) error, pollIntervalMilliseconds int64, maxInFlight int) error {
	cfg := nsq.NewConfig()
	cfg.MaxInFlight = maxInFlight
	cfg.LookupdPollInterval = time.Duration(pollIntervalMilliseconds) * time.Millisecond
	consumer, err := nsq.NewConsumer(topic, channel, cfg)
	if err != nil {
		nc.Worker().Logger().Errorf("create nsq consumer failed, err: %v", err)
		return err
	}
	concurrency := maxInFlight
	if concurrency <= 0 {
		concurrency = 1
	} else if concurrency > 100 {
		concurrency = 100
	}
	consumer.AddConcurrentHandlers(nc.nsqHandler(handler), concurrency)
	err = consumer.ConnectToNSQLookupd(nc.subServer())
	if err != nil {
		nc.Worker().Logger().Errorf("connect to nsqlookupd failed, err: %v", err)
		return err
	}
	return nil
}

func (nc *NSQClient) nsqHandler(msgHandler MsgHandler) nsq.Handler {
	return nsq.HandlerFunc(func(message *nsq.Message) error {
		return msgHandler(message.Body)
	})
}

func (nc *NSQClient) Close() {}