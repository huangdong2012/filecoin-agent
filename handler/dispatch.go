package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/sirupsen/logrus"
	"huangdong2012/filecoin-agent/filecoin"
	"huangdong2012/filecoin-agent/hlmd"
	"huangdong2012/filecoin-agent/model"
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
	logrus.Infof("recv worker sched msg: %+v ", msg)
	cmd := &model.DispatchCommand{}
	if err = json.Unmarshal([]byte(msg.Body), cmd); err != nil {
		logrus.Error(err)
		return nil, err
	}
	var (
		src        string
		dest       string
		srcClient  *filecoin.ScheduleClient
		destClient *filecoin.ScheduleClient
		repoPath   string
		api        string
		token      string
		isInit     bool
		workerId   string
		tasks      int
	)
	if destClient, err = h.BuildShedAPI(cmd.Api, cmd.Token); err != nil {
		logrus.Error(err)
		return nil, err
	}
	if dest, err = destClient.GetMinerInfo(context.TODO()); err != nil {
		logrus.Error(err)
		return nil, err
	}
	repoPath = cmd.Repo
	if repoPath == "" {
		repoPath = opt.LotusMinerConfig.RepoPath
	}
	if api, token, err = h.ReadApiToken(repoPath); err != nil {
		logrus.Error(err)
		return nil, err
	}
	isInit = api == "" && token == ""
	if isInit {
		if srcClient, err = h.BuildShedAPI(api, token); err != nil {
			logrus.Error(err)
			return nil, err
		}
		if dest, err = srcClient.GetMinerInfo(context.TODO()); err != nil {
			logrus.Error(err)
			return nil, err
		}
		if workerId, err = h.CheckIdFile(); err != nil {
			logrus.Error(err)
			return nil, err
		}
		if strings.Compare(dest, src) == 0 {
			return &model.CommandResponse{
				ID:         msg.ID,
				Host:       id,
				Status:     int(model.CommandStatus_Success),
				FinishTime: time.Now().Unix(),
				Message:    dest,
			}, nil
		}

		if err = srcClient.RequestDisableWorker(context.TODO(), workerId); err != nil {
			logrus.Error(err)
			return nil, err
		}
		for {
			tasks, err = srcClient.GetWorkerBusyTask(context.TODO(), workerId)
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
				if err = publishResp(msg.ID, model.CommandStatus_Running, fmt.Sprintf("check worker busy tasks: %v", tasks)); err != nil {
					logrus.Error(err)
				}
			}
			time.Sleep(time.Second * 10)
		}
	}

	if err = h.WriteApiToken(repoPath, cmd.Api, cmd.Token); err != nil {
		logrus.Error(err)
		return nil, err
	}

	if cmd.Verb != "" && cmd.Process != "" {
		state := map[string]string{
			"start":   "started",
			"stop":    "stopped",
			"restart": "restarted",
		}
		if _, ok := state[cmd.Verb]; ok {
			if err = h.operateServices(msg.ID, cmd.Verb, []string{cmd.Process}); err != nil {
				logrus.Error(err)
				return nil, err
			}
		}
	}
	return &model.CommandResponse{
		ID:         msg.ID,
		Host:       id,
		Status:     int(model.CommandStatus_Success),
		FinishTime: time.Now().Unix(),
		Message:    dest,
	}, nil
}

func (h *dispatchHandler) BuildShedAPI(api string, token string) (client *filecoin.ScheduleClient, err error) {
	var (
		apiEndPoint string
		ma          multiaddr.Multiaddr
		addr        string
	)
	ma, err = multiaddr.NewMultiaddr(strings.TrimSpace(api))
	if err == nil {
		_, addr, err = manet.DialArgs(ma)
		if err != nil {
			return nil, err
		}
		apiEndPoint = "http://" + addr + "/rpc/v0"
		client = filecoin.NewScheduleClient(apiEndPoint, strings.TrimSpace(token))
	}
	return
}
func (h *dispatchHandler) CheckIdFile() (id string, err error) {
	//todo test
	var (
		idFile       = "~/.lotusworker/worker.id"
		workerIdFile string
		content      []byte
	)
	workerIdFile, err = homedir.Expand(idFile)
	if err != nil {
		return "", err
	}
	// checkfile
	if _, err = os.Stat(workerIdFile); err != nil {
		if os.IsNotExist(err) {
			//直接重启
			logrus.Warnf("idfile not found: %+v", err)
			return "", nil
		} else {
			return "", err
		}
	} else {
		content, err = ioutil.ReadFile(workerIdFile)
		if err != nil {
			return "", err
		}
		id = strings.TrimSpace(string(content))
	}
	return
}
func (h *dispatchHandler) ReadApiToken(repoPath string) (api string, token string, err error) {
	var (
		tokenBytes []byte
		apiBytes   []byte
	)
	if _, err := os.Stat(repoPath); err != nil {
		if !os.IsNotExist(err) {
			return "", "", nil
		}
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			return "", "", err
		}
	}
	if _, err = os.Stat(filepath.Join(repoPath, "worker_api")); err != nil {
		return "", "", err
	}
	if _, err = os.Stat(filepath.Join(repoPath, "worker_token")); err != nil {
		return "", "", err
	}
	tokenBytes, err = ioutil.ReadFile(filepath.Join(repoPath, "worker_token"))
	if err != nil {
		return "", "", err
	}
	apiBytes, err = ioutil.ReadFile(filepath.Join(repoPath, "worker_api"))
	if err != nil {
		return "", "", err
	}
	api = strings.TrimSpace(string(apiBytes))
	token = strings.TrimSpace(string(tokenBytes))
	return
}
func (h *dispatchHandler) WriteApiToken(repoPath string, api string, token string) (err error) {
	//判断worker token 是否存在
	//1.若不存在，直接写入新的worker_api,worker_token，尝试启动worker进程
	//2.若存在,如果能连上miner，查询miner机器上worker状态，如果任务数量不为0，需要定时持续检测任务数量；如果任务数量为0，则执行1逻辑
	//  如果worker id 文件不存在,则直接重启
	if err = ioutil.WriteFile(filepath.Join(repoPath, "worker_api"), []byte(strings.TrimSpace(api)), 0600); err != nil {
		return err
	}
	if err = ioutil.WriteFile(filepath.Join(repoPath, "worker_token"), []byte(strings.TrimSpace(token)), 0600); err != nil {
		return err
	}
	return nil
}
func (h *dispatchHandler) operateServices(msgID, operate string, services []string) error {
	var (
		err  error
		flag = operate == "stop" //(stopping ~ stopped) report to kafka for progress
	)
	for _, srv := range services {
		if flag { //stopping
			if err = publishResp(msgID, model.CommandStatus_Running, fmt.Sprintf("stopping service: %v", srv)); err != nil {
				logrus.Warnf("publish stopping progress error: %+v", err.Error())
			}
		}
		if err := hlmd.Execute([]string{operate, srv}); err != nil {
			return err
		}
		if flag { //stopped
			if err = publishResp(msgID, model.CommandStatus_Running, fmt.Sprintf("stopped service: %v", srv)); err != nil {
				logrus.Warnf("publish stopped progress error: %+v", err.Error())
			}
		}
	}
	return nil
}
