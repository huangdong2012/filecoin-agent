package model

type Package struct {
	Version string `json:"version"`
	Full    bool   `json:"full"` //全量更新
	//todo...
}
