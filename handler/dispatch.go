package handler

import (
	"context"
	"encoding/json"
	"github.com/mitchellh/go-homedir"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/sirupsen/logrus"
	"huangdong2012/filecoin-agent/filecoin"
	"huangdong2012/filecoin-agent/model"
	"huangdong2012/filecoin-agent/rpcclient"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	Dispatch = &dispatchHandler{}
)

type dispatchHandler struct {
}

func (h *dispatchHandler) Handle(msg *model.CommandRequest) (resp *model.CommandResponse, err error) {
	cmd := &model.DispatchCommand{}
	if err = json.Unmarshal([]byte(msg.Body), cmd); err != nil {
		return nil, err
	}
	// token 加解密
	repoPath := cmd.Repo
	if repoPath == "" {
		repoPath = opt.LotusMinerConfig.RepoPath
	}
	if _, err := os.Stat(repoPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			return nil, err
		}
	}
	//todo test
	idFile := "~/.lotusworker/worker.id"
	workerIdFile, err := homedir.Expand(idFile)
	if err != nil {
		logrus.Infof("%+v", err)
		return nil, err
	} // checkfile
	if _, err := os.Stat(workerIdFile); err != nil {
		if os.IsNotExist(err) {
			//直接重启
			logrus.Infof("idfile not found: %+v", err)
		} else {
			logrus.Infof("read idfile err: %+v", err)
			return nil, err
		}
	} else {
		id, err := ioutil.ReadFile(workerIdFile)
		if err != nil {
			logrus.Infof("%+v", err)
			return nil, err
		}
		workerid := strings.TrimSpace(string(id))
		isInit := false
		if _, err := os.Stat(filepath.Join(repoPath, "worker_api")); err != nil {
			logrus.Infof("%+v", err)
		}
		if _, err := os.Stat(filepath.Join(repoPath, "worker_token")); err != nil {
			logrus.Infof("%+v", err)
		}
		isInit = true
		if !isInit {
			//直接重启
			logrus.Info("not init , restart quick")
		} else {
			//尝试连接miner，判断API的有效性
			minerRepo := repoPath
			token, err := ioutil.ReadFile(filepath.Join(minerRepo, "worker_token"))
			if err != nil {
				logrus.Warnf("host served as miner ? %s", err.Error())
			}
			api, err := ioutil.ReadFile(filepath.Join(minerRepo, "worker_api"))
			var apiEndPoint string
			if err != nil {
				logrus.Warnf("host served as miner ? %s", err.Error())
			} else {
				ma, err := multiaddr.NewMultiaddr(strings.TrimSpace(string(api)))
				if err == nil {
					_, addr, err := manet.DialArgs(ma)
					if err != nil {
						logrus.Infof("%+v", err)
					}
					apiEndPoint = "http://" + addr + "/rpc/v0"
				}
			}
			SchedAPI := filecoin.NewScheduleClient(apiEndPoint, strings.TrimSpace(string(token)))
			if err := SchedAPI.RequestDisableWorker(context.TODO(), workerid); err != nil {
				logrus.Infof("disable worker: %+v", err)
				return nil, err
			}
			logrus.Infof("disable worker: %+v", workerid)
			for {
				tasks, err := SchedAPI.GetWorkerBusyTask(context.TODO(), workerid)
				if err != nil {
					logrus.Infof("%+v", err)
					break
				} else {
					if tasks == 0 {
						logrus.Infof("worker has free task : %+v", tasks)
						break
					} else {
						logrus.Infof("worker busy tasks : %+v", tasks)
					}
				}
				time.Sleep(time.Second * 10)
			}
			//ready to change token
		}
	}
	//判断worker token 是否存在
	//1.若不存在，直接写入新的worker_api,worker_token，尝试启动worker进程
	//2.若存在,如果能连上miner，查询miner机器上worker状态，如果任务数量不为0，需要定时持续检测任务数量；如果任务数量为0，则执行1逻辑
	//  如果worker id 文件不存在,则直接重启
	if err = ioutil.WriteFile(filepath.Join(repoPath, "worker_api"), []byte(cmd.Api), 0600); err != nil {
		return nil, err
	}
	if err = ioutil.WriteFile(filepath.Join(repoPath, "worker_token"), []byte(cmd.Token), 0600); err != nil {
		return nil, err
	}
	//是否需要重启
	if cmd.Verb != "" && cmd.Process != "" {
		state := map[string]string{
			"start":   "started",
			"stop":    "stopped",
			"restart": "restarted",
		}
		if _, ok := state[cmd.Verb]; ok {
			hlmd := rpcclient.NewRPCClient(opt.HlmMinerConfig.ServerUrl, opt.HlmMinerConfig.UserName, opt.HlmMinerConfig.PassWord, true)
			ret, err := hlmd.ChangeProcessState(cmd.Verb, cmd.Process)
			if err != nil {
				return nil, err
			}
			logrus.Infof("%+v", ret)
		}
	}
	return &model.CommandResponse{
		ID:         msg.ID,
		Host:       id,
		Status:     int(model.CommandStatus_Success),
		FinishTime: time.Now().Unix(),
	}, nil
}
