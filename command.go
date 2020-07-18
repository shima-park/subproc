package subproc

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/docker/docker/pkg/reexec"
)

type Cmd struct {
	Cmd        *exec.Cmd
	options    CmdOptions
	pipeReader *os.File
	pipeWriter *os.File
	error      error
}

func NewCmd(cmdStr string, opts ...CmdOption) *Cmd {
	options := NewCmdOption(opts...)
	cmd := reexec.Command(append([]string{cmdStr}, options.CommandLineArguments...)...)
	cmd.Env = options.Environment
	cmd.Stdin = options.Stdin
	cmd.Stdout = options.Stdout
	cmd.Stderr = options.Stderr

	pipeReader, pipeWriter, err := os.Pipe()
	cmd.ExtraFiles = append([]*os.File{pipeWriter}, options.ExtraFiles...)

	return &Cmd{
		Cmd:        cmd,
		options:    options,
		pipeReader: pipeReader,
		pipeWriter: pipeWriter,
		error:      err,
	}
}

func (c *Cmd) Start() error {
	if c.error != nil {
		return c.error
	}

	if err := c.options.StartBefore(); err != nil {
		return err
	}

	if err := c.Cmd.Start(); err != nil {
		return err
	}

	if err := c.options.StartAfter(); err != nil {
		return err
	}

	return nil
}

func (c *Cmd) Wait() (err error) {
	if c.error != nil {
		return c.error
	}

	defer func() {
		c.pipeWriter.Close()
		c.pipeReader.Close()

		if c.options.FinalHook != nil {
			c.options.FinalHook(err)
		}
	}()

	if err = c.options.WaitBefore(); err != nil {
		return
	}

	err = c.Cmd.Wait()
	r := getResult(c.pipeReader, c.options.Response)
	if err != nil || r.Error != "" {
		if r.Error != "" {
			return fmt.Errorf("Command Error: %v, Message: %s", err, r.Error)
		} else {
			return fmt.Errorf("Command Error: %v", err)
		}
	}

	if err = c.options.WaitAfter(); err != nil {
		return
	}
	return nil
}

func (c *Cmd) Run() (err error) {
	if err = c.Start(); err != nil {
		return
	}

	return c.Wait()
}

func (c *Cmd) Stop() error {
	return c.Cmd.Process.Signal(os.Interrupt)
}

func (c *Cmd) Kill() error {
	return c.Cmd.Process.Kill()
}
