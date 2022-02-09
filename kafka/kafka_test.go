package kafka

import (
	"fmt"
	"github.com/google/uuid"
	"huangdong2012/filecoin-agent/infras"
	"huangdong2012/filecoin-agent/model"
	"testing"
	"time"
)

func setup() {
	Init(func(o *Option) {
		o.Brokers = []string{
			"localhost:9092",
			"localhost:9082",
			"localhost:9072",
		}
		o.Verbose = false
		o.Rest = false
	})
}

func TestPublish(t *testing.T) {
	setup()

	topic := "zdz.command.request"
	p, o, err := Publish(topic, infras.ToJson(&model.CommandRequest{
		ID:    uuid.New().String(),
		Kind:  int(model.CommandKind_Upgrade),
		Hosts: []string{infras.HostNo()},
		Body: infras.ToJson(&model.UpgradeCommand{
			SourceUrl: "http://localhost:81/download/lotus.tar.gz",
			Sha256:    "7be1a2f00576ad6ef8e59c874849b3e8685d71949e8d546e5854fb730de57e24",
			Services:  []string{"test"},
		}),
		CreateTime: time.Now().Unix(),
	}))
	fmt.Println(p, o, err)
}

func TestConsume(t *testing.T) {
	setup()

	c := make(chan bool)
	topic := "zdz.command.request"

	msg1, err := Consume("1", topic, c, nil)
	if err != nil {
		panic(err)
	}
	go consumeHandle("1", msg1)

	msg2, err := Consume("2", topic, c, nil)
	if err != nil {
		panic(err)
	}
	go consumeHandle("2", msg2)

	msg3, err := Consume("3", topic, c, nil)
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
