package subproc

import (
	"io"
	"os"
)

var (
	defaultCmdOptions = CmdOptions{
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		StartBefore: func() error { return nil },
		StartAfter:  func() error { return nil },
		WaitBefore:  func() error { return nil },
		WaitAfter:   func() error { return nil },
		FinalHook:   func(error) {},
	}
)

type CmdOptions struct {
	CommandLineArguments []string
	Environment          []string
	Stdin                io.Reader
	Stdout               io.Writer
	Stderr               io.Writer
	Response             interface{}
	StartBefore          func() error
	StartAfter           func() error
	WaitBefore           func() error
	WaitAfter            func() error
	FinalHook            func(error)
	ExtraFiles           []*os.File
}

type CmdOption func(*CmdOptions)

func NewCmdOption(opts ...CmdOption) CmdOptions {
	var options = defaultCmdOptions
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

func CommandLineArguments(args ...string) CmdOption {
	return func(o *CmdOptions) {
		o.CommandLineArguments = args
	}
}

func Environment(envs ...string) CmdOption {
	return func(o *CmdOptions) {
		o.Environment = envs
	}
}

func Stdin(r io.Reader) CmdOption {
	return func(o *CmdOptions) {
		o.Stdin = r
	}
}

func Stdout(w io.Writer) CmdOption {
	return func(o *CmdOptions) {
		o.Stdout = w
	}
}

func Stderr(w io.Writer) CmdOption {
	return func(o *CmdOptions) {
		o.Stderr = w
	}
}

func Response(resp interface{}) CmdOption {
	return func(o *CmdOptions) {
		o.Response = resp
	}
}

func StartBefore(f func() error) CmdOption {
	return func(o *CmdOptions) {
		o.StartBefore = f
	}
}

func StartAfter(f func() error) CmdOption {
	return func(o *CmdOptions) {
		o.StartAfter = f
	}
}

func WaitBefore(f func() error) CmdOption {
	return func(o *CmdOptions) {
		o.WaitBefore = f
	}
}

func WaitAfter(f func() error) CmdOption {
	return func(o *CmdOptions) {
		o.WaitAfter = f
	}
}

func FinalHook(f func(error)) CmdOption {
	return func(o *CmdOptions) {
		o.FinalHook = f
	}
}

func ExtraFiles(extrafiles ...*os.File) CmdOption {
	return func(o *CmdOptions) {
		o.ExtraFiles = extrafiles
	}
}
