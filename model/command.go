package model

type CommandRequest struct {
	ID         string   `json:"id"`
	User       string   `json:"user"`
	Kind       int      `json:"kind"` //enum: CommandKind
	MinerID    string   `json:"miner_id"`
	Workers    []string `json:"workers"`
	Hosts      []string `json:"hosts"`
	Body       string   `json:"body"`
	CreateTime int64    `json:"create_time"`
	ExpireTime int64    `json:"expire_time"`
}

type CommandResponse struct {
	ID         string `json:"id"`
	Host       string `json:"host"`
	Status     int    `json:"status"` //enum: CommandStatus
	Message    string `json:"message"`
	FinishTime int64  `json:"finish_time"`
}
