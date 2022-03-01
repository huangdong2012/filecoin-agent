package hlmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gwaycc/supd/config"
)

// find the supervisord.conf in following order:
// 1. $CWD/supervisord.conf
// 2. $CWD/etc/supervisord.conf
// 3. /etc/supervisord.conf
// 4. /etc/supervisor/supervisord.conf (since Supervisor 3.3.0)
// 5. ../etc/supervisord.conf (Relative to the executable)
// 6. ../supervisord.conf (Relative to the executable)
func (x *hlmdCtl) findSupervisordConf() (string, error) {
	possibleSupervisordConf := []string{
		x.opt.ConfigPath,
		os.ExpandEnv("$PRJ_ROOT/etc/supd/supd.ini"),

		"/hlm-miner/etc/supd/supd.ini",
		"/root/hlm-miner/etc/supd/supd.ini",

		"./supd.ini",
		"../supd.ini",
		"../etc/supd/supd.ini",
		"/etc/supd/supd.ini",
	}

	for _, file := range possibleSupervisordConf {
		if _, err := os.Stat(file); err == nil {
			abs_file, err := filepath.Abs(file)
			if err == nil {
				return abs_file, nil
			} else {
				return file, nil
			}
		}
	}

	return "", fmt.Errorf("fail to find supervisord.conf")
}

func (x *hlmdCtl) getServerUrl() string {
	if x.opt.ServerURL != "" {
		return strings.TrimSuffix(x.opt.ServerURL, "/")
	}

	configPath, err := x.findSupervisordConf()
	if err != nil {
		return ""
	}
	if _, err := os.Stat(configPath); err == nil {
		config := config.NewConfig(configPath)
		config.Load()
		if entry, ok := config.GetSupervisorctl(); ok {
			serverurl := entry.GetString("serverurl", "")
			if serverurl != "" {
				return serverurl
			}
		}
	}
	return "http://localhost:9002"
}

func (x *hlmdCtl) getUser() string {
	if x.opt.Username != "" {
		return x.opt.Username
	}

	configPath, err := x.findSupervisordConf()
	if err != nil {
		return ""
	}
	if _, err := os.Stat(configPath); err == nil {
		config := config.NewConfig(configPath)
		config.Load()
		if entry, ok := config.GetSupervisorctl(); ok {
			user := entry.GetString("username", "")
			return user
		}
	}
	return ""
}

func (x *hlmdCtl) getPassword() string {
	if x.opt.Password != "" {
		return x.opt.Password
	}

	configPath, err := x.findSupervisordConf()
	if err != nil {
		return ""
	}
	if _, err := os.Stat(configPath); err == nil {
		config := config.NewConfig(configPath)
		config.Load()
		if entry, ok := config.GetSupervisorctl(); ok {
			password := entry.GetString("password", "")
			return password
		}
	}
	return ""
}
