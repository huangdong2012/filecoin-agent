package main

import (
	"grandhelmsman/filecoin-agent/handler"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	handler.Init(func(opt *handler.Option) {
		opt.ProjectRoot = "/Users/huangdong/Temp/hlm-miner"

		opt.Brokers = []string{
			"localhost:9092",
			"localhost:9082",
			"localhost:9072",
		}
		opt.TopicRq = "zdz.command.request"
		opt.TopicRs = "zdz.command.response"

		opt.SupConfig = "/Users/huangdong/Temp/hlm-miner/supd/etc/supervisord.conf"
	})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, os.Interrupt)
	<-signals
	handler.Exit()
}
