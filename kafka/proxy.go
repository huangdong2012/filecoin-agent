package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"huangdong2012/filecoin-agent/model"
	"sync"
)

var (
	Normal = &normalProxy{
		producersRW: &sync.RWMutex{},
		producers:   make(map[string]sarama.SyncProducer),

		consumersRW: &sync.RWMutex{},
		consumers:   make(map[string]*syncConsumer),
	}
)

type syncConsumer struct {
	consumer sarama.Consumer
	messageC chan *model.CommandRequest
}

type normalProxy struct {
	producersRW *sync.RWMutex
	producers   map[string]sarama.SyncProducer

	consumersRW *sync.RWMutex
	consumers   map[string]*syncConsumer
}

func (p *normalProxy) Publish(topic string, value string) (int32, int64, error) {
	producer, err := p.getProducer(topic)
	if err != nil {
		return 0, 0, err
	}

	return producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(value),
	})
}

func (k *normalProxy) getProducer(topic string) (sarama.SyncProducer, error) {
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
	if producer, err = sarama.NewSyncProducer(opt.Brokers, config); err != nil {
		return nil, err
	}

	k.producers[topic] = producer
	return producer, nil
}

func (k *normalProxy) Consume(id, topic string, stopC <-chan bool, offsetOpt *OffsetOption) (<-chan *model.CommandRequest, error) {
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
	if csm, err = sarama.NewConsumer(opt.Brokers, sarama.NewConfig()); err != nil {
		return nil, err
	} else {
		if pts, err = csm.Partitions(topic); err != nil {
			return nil, err
		}
	}

	consumer = &syncConsumer{
		consumer: csm,
		messageC: make(chan *model.CommandRequest),
	}
	if err = k.initConsumer(consumer, topic, pts, stopC, offsetOpt); err != nil {
		return nil, err
	}
	k.consumers[key] = consumer
	return consumer.messageC, nil
}

func (k *normalProxy) initConsumer(kc *syncConsumer, topic string, pts []int32, stopC <-chan bool, offsetOpt *OffsetOption) error {
	pcs := make([]sarama.PartitionConsumer, 0, 0)
	for _, pt := range pts {
		offset := sarama.OffsetNewest
		if offsetOpt != nil && offsetOpt.GetOffset != nil {
			if tmp := offsetOpt.GetOffset(topic, pt); tmp >= 0 {
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
		go k.handleConsumer(kc, pc, stopC, offsetOpt)
	}

	return nil
}

func (k *normalProxy) handleConsumer(kc *syncConsumer, pc sarama.PartitionConsumer, stopC <-chan bool, offsetOpt *OffsetOption) {
	for {
		select {
		case <-stopC:
			return
		case msg := <-pc.Messages():
			cmd := &model.CommandRequest{}
			if err := json.Unmarshal(msg.Value, cmd); err == nil {
				kc.messageC <- cmd

				if offsetOpt != nil && offsetOpt.SetOffset != nil {
					offsetOpt.SetOffset(msg.Topic, msg.Partition, msg.Offset)
				}
			}
		}
	}
}
