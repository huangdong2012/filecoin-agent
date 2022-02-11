package handler

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"huangdong2012/filecoin-agent/model"
	"io/ioutil"
	"time"
)

var (
	WorkerConfig      = &workerConfigHandler{}
	WORKER_WATCH_FILE = "../../etc/worker_file.json"
)

type workerConfigHandler struct {
}

func (h *workerConfigHandler) Handle(msg *model.CommandRequest) (*model.CommandResponse, error) {
	logrus.Infof("recv worker msg: %+v ", msg)
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
	var json1 = model.WorkerConf{}
	err = json.Unmarshal(data, &json1)
	if err != nil {
		logrus.Error("json.Unmarshal_ERR::", err.Error())
	}
	logrus.Infof("read_data: %+v ", string(data))
	var str bytes.Buffer
	var workerCfg = model.WorkerConf{
		ID:                 json1.ID,
		IP:                 json1.IP,
		SvcUri:             json1.SvcUri,
		MaxTaskNum:         workerConf.MaxTaskNum,
		ParallelPledge:     workerConf.ParallelPledge,
		ParallelPrecommit1: workerConf.ParallelPreCommit1,
		ParallelPrecommit2: workerConf.ParallelPreCommit2,
		ParallelCommit:     workerConf.ParallelCommit,
		Commit2Srv:         workerConf.Commit2Srv,
		WdPoStSrv:          workerConf.WdPostSrv,
		WnPoStSrv:          workerConf.WnPostSrv,
	}
	logrus.Info("=============WorkerTaskTopic=============", workerCfg)
	byte_json, _ := json.Marshal(workerCfg)
	_ = json.Indent(&str, byte_json, "", "    ")
	var d1 = []byte(str.String())
	err2 := ioutil.WriteFile(WORKER_WATCH_FILE, d1, 0666) //写入文件(字节数组)
	if err2 != nil {
		logrus.Errorf(err2.Error())
	}
	return nil
}
