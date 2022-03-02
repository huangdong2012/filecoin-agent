package infras

import (
	"Acumes/uuid-generate/util/utils"
	"encoding/json"
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
	s := make([]string, 0)
	uuid := utils.UuidGenerate(s)
	return uuid
}

func TimeFormat(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func StringSliceContains(items []string, item string) bool {
	for _, str := range items {
		if str == item {
			return true
		}
	}
	return false
}
