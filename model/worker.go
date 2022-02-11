package model

type WorkerConf struct {
	ID                 string
	IP                 string
	SvcUri             string
	MaxTaskNum         string
	ParallelPledge     string
	ParallelPrecommit1 string
	ParallelPrecommit2 string
	ParallelCommit     string
	Commit2Srv         bool
	WdPoStSrv          bool
	WnPoStSrv          bool
}
