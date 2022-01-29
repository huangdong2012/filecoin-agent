package main

import (
	"grandhelmsman/filecoin-agent/infras"
	"grandhelmsman/filecoin-agent/supd"
	"os"
)

func main() {
	supd.Init(func(opt *supd.Option) {
		opt.ServerURL = "http://192.248.151.217:9001"
		opt.Username = "admin"
		opt.Password = "Hd19870224"
	})
	//supd.Init()
	infras.Throw(supd.Execute(os.Args[1:]))
}
