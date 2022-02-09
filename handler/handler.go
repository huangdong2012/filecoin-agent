package handler

import (
	"fmt"
	"huangdong2012/filecoin-agent/infras"
	"huangdong2012/filecoin-agent/kafka"
	"huangdong2012/filecoin-agent/model"
	"huangdong2012/filecoin-agent/supd"
	"time"
)

var (
	id    = infras.HostNo()
	exitC = make(chan bool)
	opt   = &Option{}
)

func Init(opts ...Options) {
	for _, fn := range opts {
		fn(opt)
	}

	//setup
	if !infras.PathExist(opt.ProjectRoot) {
		if opt.ProjectRoot = defaultProjectRoot(); len(opt.ProjectRoot) == 0 {
			panic("project-root invalid")
		}
	}

	//init kafka
	Offset.init()
	kafka.Init(opt.Brokers, opt.Verbose)
	if msgC, err := kafka.Consume(id, opt.TopicRq, exitC, &kafka.OffsetOption{
		GetOffset: Offset.get,
		SetOffset: Offset.set,
	}); err != nil {
		panic(err)
	} else {
		go loop(msgC)
	}

	//init supd
	if len(opt.SupConfig) > 0 {
		supd.Init(func(o *supd.Option) {
			o.ConfigPath = opt.SupConfig
			o.Verbose = opt.Verbose
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

func loop(msgC <-chan *model.CommandRequest) {
	for msg := range msgC {
		var (
			err  error
			resp *model.CommandResponse
		)
		if !infras.StringSliceContains(msg.Hosts, id) {
			continue
		}
		if msg.ExpireTime > 0 && time.Now().After(time.Unix(msg.ExpireTime, 0)) {
			continue
		}

		switch model.CommandKind(msg.Kind) {
		case model.CommandKind_Upgrade:
			resp, err = Upgrade.Handle(msg)
		}

		if err != nil {
			err = publishResp(msg.ID, model.CommandStatus_Error, err.Error())
		} else {
			err = publishResp(msg.ID, model.CommandStatus(resp.Status), resp.Message)
		}
		if err != nil {
			fmt.Println("publish command response error:", err)
		}
	}
}

func publishResp(mid string, status model.CommandStatus, msg string) error {
	_, _, err := kafka.PublishCmdResp(opt.TopicRs, &model.CommandResponse{
		ID:         mid,
		Host:       id,
		Status:     int(status),
		Message:    msg,
		FinishTime: time.Now().Unix(),
	})
	return err
}

func defaultProjectRoot() string {
	if root := "/root/hlm-miner"; infras.PathExist(root) {
		return root
	}
	if root := "/hlm-miner"; infras.PathExist(root) {
		return root
	}
	return ""
}
