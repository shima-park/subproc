package subproc

import (
	"io"
	"os"
)

var (
	defaultCmdOptions = CmdOptions{
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
		PreHook:   func() error { return nil },
		PostHook:  func() error { return nil },
		FinalHook: func(error) {},
	}
)

type CmdOptions struct {
	CommandLineArguments []string
	Environment          []string
	Stdin                io.Reader
	Stdout               io.Writer
	Stderr               io.Writer
	Response             interface{}
	PreHook              func() error
	PostHook             func() error
	FinalHook            func(error)
}

type CmdOption func(*CmdOptions)

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

func PreHook(f func() error) CmdOption {
	return func(o *CmdOptions) {
		o.PreHook = f
	}
}

func PostHook(f func() error) CmdOption {
	return func(o *CmdOptions) {
		o.PostHook = f
	}
}

func FinalHook(f func(error)) CmdOption {
	return func(o *CmdOptions) {
		o.FinalHook = f
	}
}
