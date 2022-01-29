package supd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ochinchina/supervisord/config"
)

// find the supervisord.conf in following order:
// 1. $CWD/supervisord.conf
// 2. $CWD/etc/supervisord.conf
// 3. /etc/supervisord.conf
// 4. /etc/supervisor/supervisord.conf (since Supervisor 3.3.0)
// 5. ../etc/supervisord.conf (Relative to the executable)
// 6. ../supervisord.conf (Relative to the executable)
func (x *supdCtl) findSupervisordConf() (string, error) {
	possibleSupervisordConf := []string{
		x.opt.ConfigPath,

		"/hlm-miner/etc/supd/supd.ini",
		"/root/hlm-miner/etc/supd/supd.ini",
		"/hlm-miner/etc/hlmd/hlmd.ini.ini",
		"/root/hlm-miner/etc/hlmd/hlmd.ini.ini",

		"./supervisord.conf",
		"./etc/supervisord.conf",
		"/etc/supervisord.conf",
		"/etc/supervisor/supervisord.conf",
		"../etc/supervisord.conf",
		"../supervisord.conf",
	}

	for _, file := range possibleSupervisordConf {
		if _, err := os.Stat(file); err == nil {
			absFile, err := filepath.Abs(file)
			if err == nil {
				return absFile, nil
			}
			return file, nil
		}
	}

	return "", fmt.Errorf("fail to find supervisord.conf")
}

func (x *supdCtl) getServerURL() (string, error) {
	if x.opt.ServerURL != "" {
		return strings.TrimSuffix(x.opt.ServerURL, "/"), nil
	}

	configPath, err := x.findSupervisordConf()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(configPath); err == nil {
		myconfig := config.NewConfig(configPath)
		myconfig.Load()
		if entry, ok := myconfig.GetSupervisorctl(); ok {
			serverurl := entry.GetString("serverurl", "")
			if serverurl != "" {
				return strings.TrimSuffix(serverurl, "/"), nil
			}
		}
	}
	return "http://localhost:9001", nil
}

func (x *supdCtl) getUser() (string, error) {
	if x.opt.Username != "" {
		return x.opt.Username, nil
	}

	configPath, err := x.findSupervisordConf()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(configPath); err == nil {
		myconfig := config.NewConfig(configPath)
		myconfig.Load()
		if entry, ok := myconfig.GetSupervisorctl(); ok {
			user := entry.GetString("username", "")
			return user, nil
		}
	}
	return "", errors.New("username not found")
}

func (x *supdCtl) getPassword() (string, error) {
	if x.opt.Password != "" {
		return x.opt.Password, nil
	}

	configPath, err := x.findSupervisordConf()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(configPath); err == nil {
		myconfig := config.NewConfig(configPath)
		myconfig.Load()
		if entry, ok := myconfig.GetSupervisorctl(); ok {
			password := entry.GetString("password", "")
			return password, nil
		}
	}
	return "", errors.New("password not found")
}
