package kafka

import (
	"fmt"
	"github.com/google/uuid"
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

	topic := "zdz.command.request"
	p, o, err := Publish(topic, infras.ToJson(&model.CommandRequest{
		ID:    uuid.New().String(),
		Kind:  int(model.CommandKind_Upgrade),
		Hosts: []string{infras.HostNo()},
		Body: infras.ToJson(&model.UpgradeCommand{
			SourceUrl:  "http://localhost:81/download/lotus.tar.gz",
			TargetPath: "/Users/huangdong/Temp/hlm-miner",
			Services:   []string{"test"},
		}),
		CreateTime: time.Now().Unix(),
	}))
	fmt.Println(p, o, err)
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
