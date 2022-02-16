package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"huangdong2012/filecoin-agent/model"
	"io/ioutil"
	"log"
	"time"
)

var (
	WorkerConfig      = &workerConfigHandler{}
	WORKER_WATCH_FILE = "../../etc/worker.yml"
)

type workerConfigHandler struct {
}

func (h *workerConfigHandler) Handle(msg *model.CommandRequest) (*model.CommandResponse, error) {
	err := h.handlerWorkerTask(msg)
	if err != nil {
		return nil, err
	}
	return &model.CommandResponse{
		ID:         msg.ID,
		Host:       id,
		Status:     int(model.CommandStatus_Success),
		FinishTime: time.Now().Unix(),
	}, nil
}

func (h *workerConfigHandler) handlerWorkerTask(msg *model.CommandRequest) error {
	var workerConf model.WorkerConfDto
	err := json.Unmarshal([]byte(msg.Body), &workerConf)
	if err != nil {
		return err
	}
	//先读取文件，
	data, err := ioutil.ReadFile(WORKER_WATCH_FILE)
	if err != nil {
		logrus.Error("Read_File_Err__:", err.Error())
	}
	var t = model.WorkerConf{}
	err = yaml.Unmarshal(data, &t)
	if err != nil {
		logrus.Error("json.Unmarshal_ERR::", err.Error())
	}
	logrus.Infof("read_data: %+v ", string(data))
	var str bytes.Buffer
	var workerCfg = model.WorkerConf{
		ID:                 t.ID,
		IP:                 t.IP,
		SvcUri:             t.SvcUri,
		MaxTaskNum:         workerConf.MaxTaskNum,
		ParallelPledge:     workerConf.ParallelPledge,
		ParallelPrecommit1: workerConf.ParallelPreCommit1,
		ParallelPrecommit2: workerConf.ParallelPreCommit2,
		ParallelCommit:     workerConf.ParallelCommit,
		Commit2Srv:         workerConf.Commit2Srv,
		WdPoStSrv:          workerConf.WdPostSrv,
		WnPoStSrv:          workerConf.WnPostSrv,
	}
	_ = json.Indent(&str, []byte(t.FixedEnv), "", "    ")
	workerCfg.FixedEnv = str.String()
	str.Reset()
	_ = json.Indent(&str, []byte(workerConf.EnvironmentVariable), "", "    ")
	workerCfg.EnvironmentVariable = str.String()

	logrus.Info("=============WorkerTaskTopic=============", workerCfg)
	d, err := yaml.Marshal(&t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err2 := ioutil.WriteFile(WORKER_WATCH_FILE, d, 0666) //写入文件(字节数组)
	if err2 != nil {
		fmt.Errorf(err2.Error())
	}
	return nil
}
