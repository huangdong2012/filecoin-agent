package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"huangdong2012/filecoin-agent/model"
	"sync"
)

func newNormalProxy() *normalProxy {
	return &normalProxy{
		producersRW: &sync.RWMutex{},
		producers:   make(map[string]sarama.SyncProducer),

		consumersRW: &sync.RWMutex{},
		consumers:   make(map[string]*syncConsumer),
	}
}

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

func (p *normalProxy) getProducer(topic string) (sarama.SyncProducer, error) {
	p.producersRW.RLock()
	producer, ok := p.producers[topic]
	p.producersRW.RUnlock()
	if ok {
		return producer, nil
	}

	p.producersRW.Lock()
	defer p.producersRW.Unlock()

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

	p.producers[topic] = producer
	return producer, nil
}

func (p *normalProxy) Consume(id, topic string, stopC <-chan bool, offsetOpt *OffsetOption) (<-chan *model.CommandRequest, error) {
	key := fmt.Sprintf("%v@%v", id, topic)
	p.consumersRW.RLock()
	consumer, ok := p.consumers[key]
	p.consumersRW.RUnlock()
	if ok {
		return consumer.messageC, nil
	}

	p.consumersRW.Lock()
	defer p.consumersRW.Unlock()

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
	if err = p.initConsumer(consumer, topic, pts, stopC, offsetOpt); err != nil {
		return nil, err
	}
	p.consumers[key] = consumer
	return consumer.messageC, nil
}

func (p *normalProxy) initConsumer(kc *syncConsumer, topic string, pts []int32, stopC <-chan bool, offsetOpt *OffsetOption) error {
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
		go p.handleConsumer(kc, pc, stopC, offsetOpt)
	}

	return nil
}

func (p *normalProxy) handleConsumer(kc *syncConsumer, pc sarama.PartitionConsumer, stopC <-chan bool, offsetOpt *OffsetOption) {
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
