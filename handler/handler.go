package handler

import (
	"github.com/sirupsen/logrus"
	"huangdong2012/filecoin-agent/hlmd"
	"huangdong2012/filecoin-agent/infras"
	"huangdong2012/filecoin-agent/kafka"
	"huangdong2012/filecoin-agent/model"
	"os"
	"path/filepath"
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
	kafka.Init(func(kfkOpt *kafka.Option) {
		kfkOpt.Brokers = opt.Brokers
		kfkOpt.Verbose = opt.Verbose
		kfkOpt.Rest = opt.Rest
	})
	if msgC, err := kafka.Consume(id, opt.TopicRq, exitC, &kafka.OffsetOption{
		GetOffset: Offset.get,
		SetOffset: Offset.set,
	}); err != nil {
		panic(err)
	} else {
		go loop(msgC)
	}

	//init hlmd
	if len(opt.SupConfig) > 0 {
		hlmd.Init(func(o *hlmd.Option) {
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
		logrus.Infof("command request: msg-id(%+v) msg(%+v)", msg.ID, msg)
		if !infras.StringSliceContains(msg.Hosts, id) {
			logrus.Infof("command ignore-host: msg-id(%+v) ", msg.ID)
			continue
		}
		if msg.ExpireTime > 0 && time.Now().After(time.Unix(msg.ExpireTime, 0)) {
			logrus.Infof("command ignore-expire: msg-id(%+v) ", msg.ID)
			continue
		}

		switch model.CommandKind(msg.Kind) {
		case model.CommandKind_Upgrade:
			resp, err = Upgrade.Handle(msg)
		case model.CommandKind_Dispatch:
			resp, err = Dispatch.Handle(msg)
		case model.CommandKind_WorkerConfig:
			resp, err = WorkerConfig.Handle(msg)
		}

		if err != nil {
			logrus.Errorf("command error: msg-id(%+v) error(%+v)", msg.ID, err)
			if err = publishResp(msg.ID, model.CommandStatus_Error, err.Error()); err != nil {
				logrus.Errorf("command publish error: msg-id(%+v) error(%+v)", msg.ID, err)
			}
		} else {
			logrus.Infof("command finish: msg-id(%+v)", msg.ID)
			if err = publishResp(msg.ID, model.CommandStatus(resp.Status), resp.Message); err != nil {
				logrus.Errorf("command publish error: msg-id(%+v) error(%+v)", msg.ID, err)
			}
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
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, "hlm-miner")
	}
	return os.Getenv("PRJ_ROOT")
}
