package handler

type Options func(option *Option)

type Option struct {
	ProjectRoot string //hlm-miner路径
	Verbose     bool

	Brokers []string //kafka brokers
	TopicRq string   //command-request topic
	TopicRs string   //command-response topic

	SupConfig string //supervisord config path
}
