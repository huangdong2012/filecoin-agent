package hlmd

type Options func(*Option)

type Option struct {
	ConfigPath string
	ServerURL  string
	Username   string
	Password   string
	Verbose    bool
}
