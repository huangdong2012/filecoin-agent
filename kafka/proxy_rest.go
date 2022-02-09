package kafka

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"huangdong2012/filecoin-agent/kafka/rest"
	"huangdong2012/filecoin-agent/model"
)

func newRestProxy(url string) *restProxy {
	return &restProxy{
		url: url,

		consumersRW: &sync.RWMutex{},
		consumers:   make(map[string]*restConsumer),
		publisher:   rest.NewPublisher(url),
	}
}

type restConsumer struct {
	consumer *rest.Consumer
	messageC chan *model.CommandRequest
}

type restProxy struct {
	url string

	consumersRW *sync.RWMutex
	consumers   map[string]*restConsumer
	publisher   *rest.Publisher
}

func (p *restProxy) Publish(topic string, value string) (int32, int64, error) {
	return p.publisher.Publish(topic, value)
}

func (p *restProxy) Consume(id, topic string, stopC <-chan bool, offsetOpt *OffsetOption) (<-chan *model.CommandRequest, error) {
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
		rc  *rest.Consumer
	)
	if rc, err = rest.NewConsumer(p.url, id, topic); err != nil {
		return nil, err
	}
	consumer = &restConsumer{
		consumer: rc,
		messageC: make(chan *model.CommandRequest),
	}
	go p.loopConsumer(consumer, stopC, offsetOpt)
	p.consumers[key] = consumer
	return consumer.messageC, nil
}

func (p *restProxy) loopConsumer(rc *restConsumer, stopC <-chan bool, offsetOpt *OffsetOption) {
	for {
		select {
		case <-stopC:
			return
		case <-time.After(time.Second):
			msgs, err := rc.consumer.Consume()
			if err != nil {
				fmt.Println("get messages from kafka-rest error:", err)
				continue
			}

			for _, msg := range msgs {
				cmd := &model.CommandRequest{}
				if err := json.Unmarshal([]byte(msg.Value), cmd); err == nil {
					rc.messageC <- cmd

					if offsetOpt != nil && offsetOpt.SetOffset != nil {
						offsetOpt.SetOffset(msg.Topic, msg.Partition, msg.Offset)
					}
				}
			}
		}
	}
}
