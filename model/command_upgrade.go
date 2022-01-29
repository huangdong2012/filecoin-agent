package model

type UpgradeCommand struct {
	SourceUrl  string   `json:"source_url"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	Sha256     string   `json:"sha256"`
	TargetPath string   `json:"target_path"`
	Services   []string `json:"services"`
}
