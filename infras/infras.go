package infras

import "encoding/json"

func Throw(err error) {
	if err != nil {
		panic(err)
	}
}

func ToJson(val interface{}) string {
	data, _ := json.MarshalIndent(val, "", "   ")
	return string(data)
}
