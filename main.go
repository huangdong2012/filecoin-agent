package main

import (
	"grandhelmsman/filecoin-agent/handler"
	"time"
)

func main() {
	handler.Init([]string{
		"localhost:9092",
		"localhost:9082",
		"localhost:9072",
	}, "zdz.command.request", "zdz.command.response", "", false)

	<-time.After(time.Hour) //test
}
