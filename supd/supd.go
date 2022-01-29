package supd

import (
	"grandhelmsman/filecoin-agent/infras"
)

var (
	ctl = &supdCtl{
		opt: &Option{Verbose: true},
	}
)

func Init(opts ...Options) {
	for _, fn := range opts {
		fn(ctl.opt)
	}

	var err error
	ctl.opt.ServerURL, err = ctl.getServerURL()
	infras.Throw(err)

	ctl.opt.Username, err = ctl.getUser()
	infras.Throw(err)

	ctl.opt.Password, err = ctl.getPassword()
	infras.Throw(err)
}

func Execute(args []string) error {
	return ctl.Execute(args)
}
