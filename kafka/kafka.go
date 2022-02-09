package kafka

import (
	"github.com/Shopify/sarama"
	"huangdong2012/filecoin-agent/infras"
	"huangdong2012/filecoin-agent/model"
	"log"
	"os"
)

var (
	opt   = &Option{}
	kafka Proxy
)

type Proxy interface {
	Publish(topic string, value string) (int32, int64, error)
	Consume(id, topic string, stopC <-chan bool, offsetOpt *OffsetOption) (<-chan *model.CommandRequest, error)
}

func Init(opts ...Options) {
	for _, fn := range opts {
		fn(opt)
	}
	if len(opt.Brokers) == 0 {
		panic("kafka brokers invalid")
	}
	if opt.Verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}
	if opt.Rest {
		kafka = newRestProxy(opt.Brokers[0])
	} else {
		kafka = newNormalProxy()
	}
}

func PublishCmdResp(topic string, cmdResp *model.CommandResponse) (int32, int64, error) {
	return Publish(topic, infras.ToJson(cmdResp))
}

func Publish(topic string, value string) (int32, int64, error) {
	return kafka.Publish(topic, value)
}

func Consume(id, topic string, stopC <-chan bool, offsetOpt *OffsetOption) (<-chan *model.CommandRequest, error) {
	return kafka.Consume(id, topic, stopC, offsetOpt)
}
