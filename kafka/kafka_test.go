package kafka

import (
	"fmt"
	"grandhelmsman/filecoin-agent/infras"
	"grandhelmsman/filecoin-agent/model"
	"testing"
	"time"
)

func setup() {
	Init([]string{
		"localhost:9092",
		"localhost:9082",
		"localhost:9072",
	}, false)
}

func TestPublish(t *testing.T) {
	setup()

	topic := "goto.test"
	for i := 0; i < 5; i++ {
		p, o, err := Publish(topic, infras.ToJson(&model.CommandRequest{
			ID: fmt.Sprintf("happy new year: %v", i),
		}))
		fmt.Println(p, o, err)
	}
}

func TestConsume(t *testing.T) {
	setup()

	c := make(chan bool)
	topic := "goto.test"

	msg1, err := Consume("1", topic, c)
	if err != nil {
		panic(err)
	}
	go consumeHandle("1", msg1)

	msg2, err := Consume("2", topic, c)
	if err != nil {
		panic(err)
	}
	go consumeHandle("2", msg2)

	msg3, err := Consume("3", topic, c)
	if err != nil {
		panic(err)
	}
	go consumeHandle("3", msg3)

	<-time.After(time.Second * 15)
	close(c)
}

func consumeHandle(id string, msgC <-chan *model.CommandRequest) {
	for msg := range msgC {
		fmt.Println(id, msg.ID)
	}
}
