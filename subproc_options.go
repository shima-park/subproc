package subproc

type SubProcOptions struct {
	Options []CmdOption
}

type SubProcOption func(*SubProcOptions)

func WithCmdOption(options ...CmdOption) SubProcOption {
	return func(o *SubProcOptions) {
		o.Options = options
	}
}
