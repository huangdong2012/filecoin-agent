package hlmd

var (
	ctl = &hlmdCtl{
		opt: &Option{Verbose: true},
	}
)

func Init(opts ...Options) {
	for _, fn := range opts {
		fn(ctl.opt)
	}

	ctl.opt.ServerURL = ctl.getServerUrl()
	ctl.opt.Username = ctl.getUser()
	ctl.opt.Password = ctl.getPassword()
}

func Execute(args []string) error {
	return ctl.Execute(args)
}
