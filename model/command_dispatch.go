package model

type DispatchCommand struct {
	Repo    string `json:"repo"`
	Api     string `json:"api"`
	Token   string `json:"token"`
	Verb    string `json:"verb"`
	Process string `json:"process"`
}
