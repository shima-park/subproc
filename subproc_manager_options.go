package subproc

type MatchOptions struct {
	IDs      []string
	Cmds     []string
	Status   *SubProcStatus
	HasError bool
	MatchAll bool
}

type MatchOption func(o *MatchOptions)

func NewMatchOptions(opts ...MatchOption) MatchOptions {
	var options = MatchOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	return options
}

func WithIDs(ids ...string) MatchOption {
	return func(o *MatchOptions) {
		o.IDs = ids
	}
}

func WithCmds(cmds ...string) MatchOption {
	return func(o *MatchOptions) {
		o.Cmds = cmds
	}
}

func WithStatus(status SubProcStatus) MatchOption {
	return func(o *MatchOptions) {
		o.Status = &status
	}
}

func WithHasError(hasError bool) MatchOption {
	return func(o *MatchOptions) {
		o.HasError = hasError
	}
}

func WithMatchAll() MatchOption {
	return func(o *MatchOptions) {
		o.MatchAll = true
	}
}

func (o MatchOptions) Match(sp SubProc) bool {
	if o.MatchAll {
		return true
	}

	if stringInSlice(sp.ID(), o.IDs) {
		return true
	}

	if stringInSlice(sp.Cmd(), o.Cmds) {
		return true
	}

	if o.Status != nil && *o.Status == sp.Status() {
		return true
	}

	if o.HasError && sp.Error() != nil {
		return true
	}
	return false
}

func stringInSlice(t string, slice []string) bool {
	for _, s := range slice {
		if s == t {
			return true
		}
	}
	return false
}
