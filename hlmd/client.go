package hlmd

import (
	"fmt"
	"github.com/gwaycc/supd/types"
	"github.com/gwaylib/errors"
	"os"
	"strings"
)

type hlmdCtl struct {
	opt *Option
}

func (x *hlmdCtl) Execute(args []string) error {
	if len(args) == 0 {
		return nil
	}

	rpcc := x.createRpcClient()
	verb := args[0]

	switch verb {

	////////////////////////////////////////////////////////////////////////////////
	// STATUS
	////////////////////////////////////////////////////////////////////////////////
	case "status":
		x.status(rpcc, args[1:])

		////////////////////////////////////////////////////////////////////////////////
		// START or STOP
		////////////////////////////////////////////////////////////////////////////////
	case "start", "stop", "restart":
		return x.startStopProcesses(rpcc, verb, args[1:])

		////////////////////////////////////////////////////////////////////////////////
		// SHUTDOWN
		////////////////////////////////////////////////////////////////////////////////
	case "shutdown":
		x.shutdown(rpcc)
	case "reload":
		x.reload(rpcc)
	case "set-env":
		x.setEnv(rpcc, args[1:])
	case "get-env":
		x.getEnv(rpcc, args[1:])
	case "signal":
		sig_name, processes := args[1], args[2:]
		x.signal(rpcc, sig_name, processes)
	case "pid":
		x.getPid(rpcc, args[1])
	default:
		fmt.Println("unknown command")
	}

	return nil
}

func (x *hlmdCtl) createRpcClient() *RPCClient {
	rpcc := NewRPCClient(x.getServerUrl(), x.getUser(), x.getPassword(), x.opt.Verbose)
	return rpcc
}

// get the status of processes
func (x *hlmdCtl) status(rpcc *RPCClient, processes []string) {
	processesMap := make(map[string]bool)
	for _, process := range processes {
		processesMap[process] = true
	}
	ret, err := rpcc.GetAllProcessInfo()
	if err != nil {
		fmt.Println(errors.As(err))
		os.Exit(1)
		return
	}
	x.showProcessInfo(ret.AllProcessInfo, processesMap)
}

// start or stop the processes
// verb must be: start or stop
func (x *hlmdCtl) startStopProcesses(rpcc *RPCClient, verb string, processes []string) error {
	state := map[string]string{
		"start":   "started",
		"stop":    "stopped",
		"restart": "restarted",
	}
	return x._startStopProcesses(rpcc, verb, processes, state[verb], true)
}

func (x *hlmdCtl) _startStopProcesses(rpcc *RPCClient, verb string, processes []string, state string, showProcessInfo bool) error {
	if len(processes) <= 0 {
		return fmt.Errorf("Please specify process for %s\n", verb)
	}
	for _, pname := range processes {
		if pname == "all" {
			if reply, err := rpcc.ChangeAllProcessState(verb); err != nil {
				return fmt.Errorf("Fail to change all process state to %s(%v)", state, err)
			} else {
				if showProcessInfo {
					x.showProcessInfo(reply.AllProcessInfo, make(map[string]bool))
				}
				return nil
			}
		} else {
			if reply, err := rpcc.ChangeProcessState(verb, pname); err != nil {
				return fmt.Errorf("%s: failed [%v]\n", pname, err)
			} else {
				if !reply.Success {
					return fmt.Errorf("%s: not %s\n", pname, state)
				}
				if showProcessInfo {
					fmt.Printf("%s: %s\n", pname, state)
				}
			}
		}
	}
	return nil
}

func (x *hlmdCtl) restartProcesses(rpcc *RPCClient, processes []string) error {
	return x._startStopProcesses(rpcc, "restart", processes, "restarted", true)
}

// shutdown the supervisord
func (x *hlmdCtl) shutdown(rpcc *RPCClient) {
	if reply, err := rpcc.Shutdown(); err == nil {
		if reply.Success {
			fmt.Printf("Shut Down\n")
		} else {
			fmt.Printf("Hmmm! Something gone wrong?!\n")
		}
	} else {
		os.Exit(1)
	}
}

// reload all the programs in the supervisord
func (x *hlmdCtl) reload(rpcc *RPCClient) {
	if reply, err := rpcc.ReloadConfig(); err == nil {

		if len(reply.AddedGroup) > 0 {
			fmt.Printf("Added Groups: %s\n", strings.Join(reply.AddedGroup, ","))
		}
		if len(reply.ChangedGroup) > 0 {
			fmt.Printf("Changed Groups: %s\n", strings.Join(reply.ChangedGroup, ","))
		}
		if len(reply.RemovedGroup) > 0 {
			fmt.Printf("Removed Groups: %s\n", strings.Join(reply.RemovedGroup, ","))
		}
	} else {
		os.Exit(1)
	}
}

func (x *hlmdCtl) setEnv(rpcc *RPCClient, args []string) {
	if len(args) < 2 {
		fmt.Println("need two args for [key value]")
		return
	}
	if _, err := rpcc.SetEnv(&SetEnvArg{Key: args[0], Value: args[1]}); err != nil {
		fmt.Println(err.Error())
		return
	}
}
func (x *hlmdCtl) getEnv(rpcc *RPCClient, args []string) {
	if len(args) < 1 {
		fmt.Println("need env key")
		return
	}
	ret, err := rpcc.GetEnv(&GetEnvArg{Key: args[0]})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(ret.Value)
}

// send signal to one or more processes
func (x *hlmdCtl) signal(rpcc *RPCClient, sig_name string, processes []string) {
	for _, process := range processes {
		if process == "all" {
			reply, err := rpcc.SignalAllProcesses(&SignalAllProcessesArg{
				Signal: sig_name,
			})
			if err == nil {
				x.showProcessInfo(reply.AllProcessInfo, make(map[string]bool))
			} else {
				fmt.Printf("Fail to send signal %s to all process", sig_name)
				os.Exit(1)
			}
		} else {
			reply, err := rpcc.SignalProcess(&SignalProcessArg{
				ProcName: process,
				Signal:   sig_name,
			})
			if err == nil && reply.Success {
				fmt.Printf("Succeed to send signal %s to process %s\n", sig_name, process)
			} else {
				fmt.Printf("Fail to send signal %s to process %s\n", sig_name, process)
				os.Exit(1)
			}
		}
	}
}

// get the pid of running program
func (x *hlmdCtl) getPid(rpcc *RPCClient, process string) {
	ret, err := rpcc.GetProcessInfo(&GetProcessInfoArg{process})
	if err != nil {
		fmt.Printf("program '%s' not found\n", process)
		os.Exit(1)
		return
	}
	fmt.Printf("%d\n", ret.ProcessInfo.Pid)
}

// check if group name should be displayed
func (x *hlmdCtl) showGroupName() bool {
	val, ok := os.LookupEnv("SUPERVISOR_GROUP_DISPLAY")
	if !ok {
		return false
	}

	val = strings.ToLower(val)
	return val == "yes" || val == "true" || val == "y" || val == "t" || val == "1"
}

func (x *hlmdCtl) showProcessInfo(allInfo []types.ProcessInfo, processesMap map[string]bool) {
	for _, pinfo := range allInfo {
		description := pinfo.Description
		if x.inProcessMap(&pinfo, processesMap) {
			processName := pinfo.GetFullName()
			if !x.showGroupName() {
				processName = pinfo.Name
			}
			fmt.Printf("%s%-33s %-10s%s%s\n", x.getANSIColor(pinfo.Statename), processName, pinfo.Statename, description, "\x1b[0m")
		}
	}
}

func (x *hlmdCtl) inProcessMap(procInfo *types.ProcessInfo, processesMap map[string]bool) bool {
	if len(processesMap) <= 0 {
		return true
	}
	for procName, _ := range processesMap {
		if procName == procInfo.Name || procName == procInfo.GetFullName() {
			return true
		}

		// check the wildcast '*'
		pos := strings.Index(procName, ":")
		if pos != -1 {
			groupName := procName[0:pos]
			programName := procName[pos+1:]
			if programName == "*" && groupName == procInfo.Group {
				return true
			}
		}
	}
	return false
}

func (x *hlmdCtl) getANSIColor(statename string) string {
	if statename == "RUNNING" {
		// green
		return "\x1b[0;32m"
	} else if statename == "BACKOFF" || statename == "FATAL" {
		// red
		return "\x1b[0;31m"
	} else {
		// yellow
		return "\x1b[1;33m"
	}
}
