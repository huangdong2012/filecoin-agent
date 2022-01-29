package main

import (
	"fmt"
	"grandhelmsman/filecoin-agent/infras"
	"grandhelmsman/filecoin-agent/kafka"
	"grandhelmsman/filecoin-agent/model"
	"time"
)

func main() {
	kafka.Init([]string{
		"localhost:9092",
		"localhost:9082",
		"localhost:9072",
	}, true)
	for i := 0; i < 5; i++ {
		p, o, err := kafka.Publish("goto.test", infras.ToJson(&model.CommandRequest{
			ID: fmt.Sprintf("haha: %v", i),
		}))
		fmt.Println(p, o, err)
	}

	c := make(chan bool)
	msg1, err := kafka.Consume("1", "goto.test", c)
	if err != nil {
		panic(err)
	}
	go handle("1", msg1)

	msg2, err := kafka.Consume("2", "goto.test", c)
	if err != nil {
		panic(err)
	}
	go handle("2", msg2)

	msg3, err := kafka.Consume("3", "goto.test", c)
	if err != nil {
		panic(err)
	}
	go handle("3", msg3)

	<-time.After(time.Second * 15)
	close(c)

	//supd.Init(func(opt *supd.Option) {
	//	opt.ServerURL = "http://192.248.151.217:9001"
	//	opt.Username = "admin"
	//	opt.Password = "Hd19870224"
	//})
	////supd.Init()
	//infras.Throw(supd.Execute(os.Args[1:]))
}

func handle(id string, msgC <-chan *model.CommandRequest) {
	for msg := range msgC {
		fmt.Println(id, msg.ID)
	}
}
