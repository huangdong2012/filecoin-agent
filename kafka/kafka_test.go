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
		//o.Brokers = []string{
		//	"localhost:9092",
		//	"localhost:9082",
		//	"localhost:9072",
		//}
		//o.Rest = false

		o.Brokers = []string{
			"http://103.44.247.17:28082",
		}
		o.Rest = true
	})
}

func TestPublish(t *testing.T) {
	setup()

	topic := "zdz.command.request"
	p, o, err := Publish(topic, infras.ToJson(&model.CommandRequest{
		ID:    uuid.New().String(),
		Kind:  int(model.CommandKind_WorkerConfig),
		Hosts: []string{"7c7d4a8d-9f69-14d5-a899-a85e455acaa1"},
		Body: infras.ToJson(&model.WorkerConfDto{
			MaxTaskNum:         10,
			ParallelPledge:     10,
			ParallelPreCommit1: 10,
			ParallelPreCommit2: 10,
			ParallelCommit:     10,
			Commit2Srv:         false,
			WdPostSrv:          false,
			WnPostSrv:          false,
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

	<-time.After(time.Second * 60)
	close(c)
}

func consumeHandle(id string, msgC <-chan *model.CommandRequest) {
	for msg := range msgC {
		fmt.Println(id, msg.ID)
	}
}
