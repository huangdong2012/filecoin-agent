package model

type UpgradeCommand struct {
	SourceUrl string   `json:"source_url"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Sha256    string   `json:"sha256"`
	Services  []string `json:"services"`
}
