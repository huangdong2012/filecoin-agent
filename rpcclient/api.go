package rpcclient

import (
	"fmt"
	"github.com/gwaylib/errors"
)

type ProcessInfo struct {
	Name          string `xml:"name" json:"name"`
	Group         string `xml:"group" json:"group"`
	Description   string `xml:"description" json:"description"`
	Start         int    `xml:"start" json:"start"`
	Stop          int    `xml:"stop" json:"stop"`
	Now           int    `xml:"now" json:"now"`
	State         int    `xml:"state" json:"state"`
	Statename     string `xml:"statename" json:"statename"`
	Spawnerr      string `xml:"spawnerr" json:"spawnerr"`
	Exitstatus    int    `xml:"exitstatus" json:"exitstatus"`
	Logfile       string `xml:"logfile" json:"logfile"`
	StdoutLogfile string `xml:"stdout_logfile" json:"stdout_logfile"`
	StderrLogfile string `xml:"stderr_logfile" json:"stderr_logfile"`
	Pid           int    `xml:"pid" json:"pid"`
	Directory     string `xml:"directory" json:"directory"`
	Command       string `xml:"directory" json:"command"`
	IniPath       string `xml:"ini_path" json:"ini_path"`
}

type ReloadConfigResult struct {
	AddedGroup   []string
	ChangedGroup []string
	RemovedGroup []string
}

type ProcessSignal struct {
	Name   string
	Signal string
}

type BooleanReply struct {
	Success bool
}

func (pi ProcessInfo) GetFullName() string {
	if len(pi.Group) > 0 {
		return fmt.Sprintf("%s:%s", pi.Group, pi.Name)
	} else {
		return pi.Name
	}
}

type StatusReply struct {
	Success bool
}

type ProcessInfoReply struct {
	ProcessInfo *ProcessInfo
}

type AllProcessInfoReply struct {
	AllProcessInfo []ProcessInfo
}

type GetVersionArg struct {
}
type GetVersionRet struct {
	Version string
}

func (r *RPCClient) GetVersion() (*GetVersionRet, error) {
	in := &GetVersionArg{}
	ret := &GetVersionRet{}
	if err := r.call("Supervisor.GetVersion", in, ret); err != nil {
		return nil, errors.As(err)
	}
	return ret, nil
}

type GetAllProcessInfoArg struct {
}
type GetAllProcessInfoRet AllProcessInfoReply

func (r *RPCClient) GetAllProcessInfo() (*GetAllProcessInfoRet, error) {
	in := &GetAllProcessInfoArg{}
	ret := &GetAllProcessInfoRet{}
	if err := r.call("Supervisor.GetAllProcessInfo", in, ret); err != nil {
		return nil, errors.As(err)
	}
	return ret, nil
}

func getChangeName(change string) (string, error) {
	srvName := ""
	switch change {
	case "start":
		srvName = "Start"
	case "stop":
		srvName = "Stop"
	case "restart":
		srvName = "Restart"
	default:
		return "", errors.New("Incorrect required state")
	}
	return srvName, nil
}

type ChangeProcessStateArg struct {
	Name string
}

type ChangeProcessStateRet struct {
	Success bool
}

func (r *RPCClient) ChangeProcessState(change string, processName string) (*ChangeProcessStateRet, error) {
	srvName, err := getChangeName(change)
	if err != nil {
		return nil, errors.As(err)
	}

	in := &ChangeProcessStateArg{processName}
	ret := &ChangeProcessStateRet{}
	if err := r.call(fmt.Sprintf("Supervisor.%sProcess", srvName), in, ret); err != nil {
		return nil, errors.As(err)
	}
	return ret, nil
}

type ChangeAllProcessStateArg struct {
	Wait bool
}
type ChangeAllProcessStateRet AllProcessInfoReply

func (r *RPCClient) ChangeAllProcessState(change string) (*ChangeAllProcessStateRet, error) {
	srvName, err := getChangeName(change)
	if err != nil {
		return nil, errors.As(err)
	}
	in := &ChangeAllProcessStateArg{true}
	ret := &ChangeAllProcessStateRet{}
	if err := r.call(fmt.Sprintf("Supervisor.%sAllProcesses", srvName), in, ret); err != nil {
		return nil, errors.As(err)
	}
	return ret, nil
}

type ShutdownArg struct {
}
type ShutdownRet struct {
	Success bool
}

func (r *RPCClient) Shutdown() (*ShutdownRet, error) {
	in := &ShutdownArg{}
	ret := &ShutdownRet{}
	if err := r.call("Supervisor.Shutdown", in, ret); err != nil {
		return nil, errors.As(err)
	}
	return ret, nil
}

type ReloadConfigArg struct {
}
type ReloadConfigRet ReloadConfigResult

func (r *RPCClient) ReloadConfig() (*ReloadConfigRet, error) {
	in := &ReloadConfigArg{}
	ret := &ReloadConfigRet{}
	ret.AddedGroup = make([]string, 0)
	ret.ChangedGroup = make([]string, 0)
	ret.RemovedGroup = make([]string, 0)
	if err := r.call("Supervisor.ReloadConfig", in, ret); err != nil {
		return nil, errors.As(err)
	}
	return ret, nil
}

type SignalProcessArg struct {
	ProcName string
	Signal   string
}

type SignalProcessRet BooleanReply

func (r *RPCClient) SignalProcess(in *SignalProcessArg) (*SignalProcessRet, error) {
	ret := &SignalProcessRet{}
	if err := r.call("Supervisor.SignalProcess", &in, ret); err != nil {
		return nil, errors.As(err)
	}
	return ret, nil
}

type SignalAllProcessesArg struct {
	Signal string
}
type SignalAllProcessesRet AllProcessInfoReply

func (r *RPCClient) SignalAllProcesses(in *SignalAllProcessesArg) (*SignalAllProcessesRet, error) {
	ret := &SignalAllProcessesRet{}
	if err := r.call("Supervisor.SignalAllProcesses", in, ret); err != nil {
		return nil, errors.As(err)
	}
	return ret, nil
}

type GetProcessInfoArg struct {
	Name string
}
type GetProcessInfoRet ProcessInfoReply

func (r *RPCClient) GetProcessInfo(in *GetProcessInfoArg) (*GetProcessInfoRet, error) {
	ret := &GetProcessInfoRet{}
	if err := r.call("Supervisor.GetProcessInfo", in, ret); err != nil {
		return nil, errors.As(err)
	}
	return ret, nil
}

type SetEnvArg struct {
	Key   string
	Value string
}
type SetEnvRet struct {
}

func (r *RPCClient) SetEnv(in *SetEnvArg) (*SetEnvRet, error) {
	ret := &SetEnvRet{}
	if err := r.call("Supervisor.SetEnv", in, ret); err != nil {
		return nil, errors.As(err)
	}
	return ret, nil
}

type GetEnvArg struct {
	Key string
}
type GetEnvRet struct {
	Value string
}

func (r *RPCClient) GetEnv(in *GetEnvArg) (*GetEnvRet, error) {
	ret := &GetEnvRet{}
	if err := r.call("Supervisor.GetEnv", in, ret); err != nil {
		return nil, errors.As(err)
	}
	return ret, nil
}
