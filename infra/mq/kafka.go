package mqclient

import (
	dt "DT-Go"
	"DT-Go/utils"
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/avast/retry-go"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

const (
	Plain       = "PLAIN"
	ScramSHA256 = "SCRAM-SHA-256"
	ScramSHA512 = "SCRAM-SHA-512"
)

type KafkaClient struct {
	dt.Infra
	username string
	password string
	brokers  []string

	mechanismProtocol string //`plain` or `scram-sha-256` or `scram-sha-512`
	saslMechanism     sasl.Mechanism
	tlsConfig         *tls.Config
}

func NewKafkaClient(pubServer string, pubPort int, subServer string, subPort int) MQClient {
	addrs := strings.Split(strings.TrimSpace(pubServer), ",")
	brokers := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		brokers = append(brokers, fmt.Sprintf("%s:%d", utils.ParseHost(addr), pubPort))
	}

	return &KafkaClient{
		brokers: brokers,
	}
}

func (kc *KafkaClient) initialize() (err error) {
	if kc.saslMechanism != nil {
		return
	}
	var m sasl.Mechanism
	switch kc.mechanismProtocol {
	case ScramSHA256:
		m, err = scram.Mechanism(scram.SHA256, kc.username, kc.password)
		if err != nil {
			return
		}
	case ScramSHA512:
		m, err = scram.Mechanism(scram.SHA512, kc.username, kc.password)
		if err != nil {
			return
		}
	case Plain:
		m = plain.Mechanism{Username: kc.username, Password: kc.password}
	default:
	}
	kc.saslMechanism = m
	return
}

func (kc *KafkaClient) Pub(topic string, msg []byte) (err error) {
	if err = kc.initialize(); err != nil {
		kc.Worker().Logger().Errorf("initialize kafka client failed, err: %v", err)
		return
	}
	w := &kafka.Writer{
		Addr:  kafka.TCP(kc.brokers...),
		Topic: topic,
		Transport: &kafka.Transport{
			TLS:  kc.tlsConfig,
			SASL: kc.saslMechanism,
		},
		AllowAutoTopicCreation: true,
	}
	defer w.Close()

	maxAttempts := uint(200)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return retry.Do(
		func() error {
			return w.WriteMessages(ctx, kafka.Message{Value: msg})
		},
		retry.Attempts(maxAttempts),
		retry.Delay(500*time.Millisecond),
		retry.OnRetry(func(n uint, err error) {
			if n > 0 {
				kc.Worker().Logger().Errorf("retrying to publish message: %v", err)
			}
		}),
		retry.RetryIf(func(err error) bool {
			return err != nil
		}),
		retry.MaxDelay(5*time.Second),
		retry.Context(ctx),
		retry.LastErrorOnly(true),
	)
}

func (kc *KafkaClient) Sub(topic string, channel string, handler func([]byte) error, pollIntervalMilliseconds int64, maxInFlight int) (err error) {
	if err = kc.initialize(); err != nil {
		kc.Worker().Logger().Errorf("initialize kafka client failed, err: %v", err)
		return
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        kc.brokers,
		GroupID:        channel,
		Topic:          topic,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        time.Duration(pollIntervalMilliseconds) * time.Millisecond,
		CommitInterval: 1 * time.Second,
		Dialer: &kafka.Dialer{
			TLS:           kc.tlsConfig,
			SASLMechanism: kc.saslMechanism,
			Timeout:       10 * time.Second,
		},
	})
	defer r.Close()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for {
			m, err := r.FetchMessage(context.Background())
			if err != nil {
				kc.Worker().Logger().Errorf("fetch message failed, err: %v", err)
				continue
			}
			if err = handler(m.Value); err != nil {
				if err = r.CommitMessages(context.Background(), m); err != nil {
					kc.Worker().Logger().Errorf("commit message failed, err: %v", err)
				}
			}
		}
	}()
	<-sigChan
	kc.Worker().Logger().Infof("wait for consumer completed...")
	return nil
}

func (kc *KafkaClient) Close() {}
