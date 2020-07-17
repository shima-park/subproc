package subproc

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/docker/docker/pkg/reexec"
)

type Cmd struct {
	*exec.Cmd
	options CmdOptions
}

func NewCmd(cmdStr string, opts ...CmdOption) *Cmd {
	var options = defaultCmdOptions
	for _, opt := range opts {
		opt(&options)
	}

	cmd := &Cmd{
		Cmd:     reexec.Command(append([]string{cmdStr}, options.CommandLineArguments...)...),
		options: options,
	}
	cmd.Env = options.Environment
	cmd.Stdin = options.Stdin
	cmd.Stdout = options.Stdout
	cmd.Stderr = options.Stderr

	return cmd
}

func (c *Cmd) Run() error {
	return c.run()
}

func (c *Cmd) run() (err error) {
	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		return err
	}

	c.ExtraFiles = []*os.File{
		pipeWriter,
	}

	defer func() {
		pipeWriter.Close()
		pipeReader.Close()

		if c.options.FinalHook != nil {
			c.options.FinalHook(err)
		}
	}()

	if c.options.PreHook != nil {
		if err = c.options.PreHook(); err != nil {
			return
		}
	}

	err = c.Cmd.Run()
	if err != nil {
		r, resErr := getResult(pipeReader)
		if resErr != nil {
			return fmt.Errorf("Command Error: %v, Failed to get result: %v", err, resErr)
		}
		if r.Error != "" {
			return fmt.Errorf("Command Error: %v, Message: %s", err, r.Error)
		}
		return fmt.Errorf("Command Error: %v", err)
	}

	if c.options.PostHook != nil {
		return c.options.PostHook()
	}
	return nil
}
