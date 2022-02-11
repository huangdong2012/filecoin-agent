package model

type WorkerConf struct {
	ID                 string
	IP                 string
	SvcUri             string
	MaxTaskNum         int
	ParallelPledge     int
	ParallelPrecommit1 int
	ParallelPrecommit2 int
	ParallelCommit     int
	Commit2Srv         bool
	WdPoStSrv          bool
	WnPoStSrv          bool
}

type WorkerConfDto struct {
	MaxTaskNum         int  `json:"max_task_num"`
	ParallelPledge     int  `json:"parallel_pledge"`
	ParallelPreCommit1 int  `json:"parallel_precommit1"`
	ParallelPreCommit2 int  `json:"parallel_precommit2"`
	ParallelCommit     int  `json:"parallel_commit"`
	Commit2Srv         bool `json:"commit2_srv"`
	WdPostSrv          bool `json:"wd_post_srv"`
	WnPostSrv          bool `json:"wn_post_srv"`
}
