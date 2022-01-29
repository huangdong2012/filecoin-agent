package model

type CommandKind int

const (
	CommandKind_None CommandKind = iota
	CommandKind_Config
	CommandKind_Dispatch
	CommandKind_Upgrade
	CommandKind_Pledge
	CommandKind_WorkerStatus
	CommandKind_WorkerProcess
)

type CommandStatus int

const (
	CommandStatus_Running CommandStatus = iota
	CommandStatus_Success
	CommandStatus_Error
)
