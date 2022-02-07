package infras

import (
	"encoding/json"
	"github.com/shirou/gopsutil/host"
	"time"
)

func Throw(err error) {
	if err != nil {
		panic(err)
	}
}

func ToJson(val interface{}) string {
	data, _ := json.MarshalIndent(val, "", "   ")
	return string(data)
}

func HostNo() string {
	info, err := host.Info()
	if err != nil {
		return ""
	}
	return info.HostID
}

func TimeFormat(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
