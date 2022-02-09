package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"huangdong2012/filecoin-agent/model"
	"sync"
)

func newProxy() *kafkaProxy {
	return &kafkaProxy{
		consumersRW: &sync.RWMutex{},
		consumers:   make(map[string]*kafkaConsumer),

		producersRW: &sync.RWMutex{},
		producers:   make(map[string]sarama.SyncProducer),
	}
}

type OffsetOption struct {
	GetOffset func(partition int32) int64
	SetOffset func(partition int32, offset int64)
}

type kafkaConsumer struct {
	consumer sarama.Consumer
	messageC chan *model.CommandRequest
}

type kafkaProxy struct {
	brokers []string

	consumersRW *sync.RWMutex
	consumers   map[string]*kafkaConsumer

	producersRW *sync.RWMutex
	producers   map[string]sarama.SyncProducer
}

func (k *kafkaProxy) getProducer(topic string) (sarama.SyncProducer, error) {
	k.producersRW.RLock()
	producer, ok := k.producers[topic]
	k.producersRW.RUnlock()
	if ok {
		return producer, nil
	}

	k.producersRW.Lock()
	defer k.producersRW.Unlock()

	var (
		err    error
		config = sarama.NewConfig()
	)
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewHashPartitioner
	if producer, err = sarama.NewSyncProducer(k.brokers, config); err != nil {
		return nil, err
	}

	k.producers[topic] = producer
	return producer, nil
}

func (k *kafkaProxy) getConsumer(id, topic string, stopC <-chan bool, opt *OffsetOption) (<-chan *model.CommandRequest, error) {
	key := fmt.Sprintf("%v@%v", id, topic)
	k.consumersRW.RLock()
	consumer, ok := k.consumers[key]
	k.consumersRW.RUnlock()
	if ok {
		return consumer.messageC, nil
	}

	k.consumersRW.Lock()
	defer k.consumersRW.Unlock()

	var (
		err error
		csm sarama.Consumer
		pts []int32
	)
	if csm, err = sarama.NewConsumer(k.brokers, sarama.NewConfig()); err != nil {
		return nil, err
	} else {
		if pts, err = csm.Partitions(topic); err != nil {
			return nil, err
		}
	}

	consumer = &kafkaConsumer{
		consumer: csm,
		messageC: make(chan *model.CommandRequest),
	}
	if err = k.initConsumer(consumer, topic, pts, stopC, opt); err != nil {
		return nil, err
	}
	k.consumers[key] = consumer
	return consumer.messageC, nil
}

func (k *kafkaProxy) initConsumer(kc *kafkaConsumer, topic string, pts []int32, stopC <-chan bool, opt *OffsetOption) error {
	pcs := make([]sarama.PartitionConsumer, 0, 0)
	for _, pt := range pts {
		offset := sarama.OffsetNewest
		if opt != nil && opt.GetOffset != nil {
			if tmp := opt.GetOffset(pt); tmp >= 0 {
				offset = tmp + 1
			}
		}
		pc, err := kc.consumer.ConsumePartition(topic, pt, offset)
		if err != nil {
			return err
		}
		pcs = append(pcs, pc)
	}
	for _, pc := range pcs {
		go k.handleConsumer(kc, pc, stopC, opt)
	}

	return nil
}

func (k *kafkaProxy) handleConsumer(kc *kafkaConsumer, pc sarama.PartitionConsumer, stopC <-chan bool, opt *OffsetOption) {
	for {
		select {
		case <-stopC:
			return
		case msg := <-pc.Messages():
			cmd := &model.CommandRequest{}
			if err := json.Unmarshal(msg.Value, cmd); err == nil {
				kc.messageC <- cmd
				if opt != nil && opt.SetOffset != nil {
					opt.SetOffset(msg.Partition, msg.Offset)
				}
			}
		}
	}
}
