package infras

import (
	"encoding/json"
	"github.com/shirou/gopsutil/host"
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
