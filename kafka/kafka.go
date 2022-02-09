package kafka

import (
	"github.com/Shopify/sarama"
	"huangdong2012/filecoin-agent/infras"
	"huangdong2012/filecoin-agent/model"
	"log"
	"os"
)

var (
	kafka = newProxy()
)

func Init(brokers []string, verbose bool) {
	if kafka.brokers = brokers; len(brokers) == 0 {
		panic("brokers invalid")
	}
	if verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}
}

func PublishCmdResp(topic string, cmdResp *model.CommandResponse) (int32, int64, error) {
	return Publish(topic, infras.ToJson(cmdResp))
}

func Publish(topic string, value string) (int32, int64, error) {
	producer, err := kafka.getProducer(topic)
	if err != nil {
		return 0, 0, err
	}

	return producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(value),
	})
}

func Consume(id, topic string, stopC <-chan bool, opt *OffsetOption) (<-chan *model.CommandRequest, error) {
	return kafka.getConsumer(id, topic, stopC, opt)
}
