package handler

type Options func(option *Option)

type Option struct {
	ProjectRoot string //hlm-miner路径
	Verbose     bool

	Brokers []string //kafka brokers
	TopicRq string   //command-request topic
	TopicRs string   //command-response topic

	SupConfig string //supervisord config path

	HlmMinerConfig HlmMiner // hlmd rpc
	LotusMinerConfig LotusMiner
}
type HlmMiner struct {
	ServerUrl string `json:"serverurl"`
	UserName  string `json:"username"`
	PassWord  string `json:"password"`
}

type LotusMiner struct {
	RepoPath   string `json:"repo"`
	Enabled    bool   `json:"enabled"`
	TimingTask int    `json:"timingTask"`
	MinerId    string `json:"minerId"`
	IsTest     bool   `json:"isTest"`
	IDFile     string `json:"idFile"`
}