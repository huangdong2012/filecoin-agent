package handler

import (
	"fmt"
	"grandhelmsman/filecoin-agent/infras"
	"grandhelmsman/filecoin-agent/kafka"
	"grandhelmsman/filecoin-agent/model"
	"grandhelmsman/filecoin-agent/supd"
	"time"
)

var (
	id    = infras.HostNo()
	exitC = make(chan bool)
)

func Init(brokers []string, topicRq, topicRs string, supConf string, verbose bool) {
	//init kafka
	kafka.Init(brokers, verbose)
	if msgC, err := kafka.Consume(id, topicRq, exitC); err != nil {
		panic(err)
	} else {
		go loop(topicRs, msgC)
	}

	//init supd
	if len(supConf) > 0 {
		supd.Init(func(opt *supd.Option) {
			opt.ConfigPath = supConf
			opt.Verbose = verbose
		})
	}
}

func Exit() {
	select {
	case <-exitC:
	default:
		close(exitC)
	}
}

func loop(topicRs string, msgC <-chan *model.CommandRequest) {
	for msg := range msgC {
		var (
			err  error
			resp *model.CommandResponse
		)
		switch model.CommandKind(msg.Kind) {
		case model.CommandKind_Upgrade:
			resp, err = Upgrade.Handle(msg)
		}
		if err != nil {
			resp = &model.CommandResponse{
				ID:         msg.ID,
				Host:       infras.HostNo(),
				Status:     int(model.CommandStatus_Error),
				Message:    err.Error(),
				FinishTime: time.Now().Unix(),
			}
		}
		if _, _, err = kafka.PublishCmdResp(topicRs, resp); err != nil {
			fmt.Println("publish command response error:", err)
		}
	}
}
