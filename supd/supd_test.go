package supd

import (
	"grandhelmsman/filecoin-agent/infras"
	"os"
	"testing"
)

func TestExecute(t *testing.T) {
	os.Args = []string{
		"main",
		"start", //"stop",
		"test",
	}

	Init(func(opt *Option) {
		opt.ServerURL = "http://155.138.158.144:9001"
		opt.Username = "admin"
		opt.Password = "Hd19870224"
	})
	//Init()
	infras.Throw(Execute(os.Args[1:]))
}
