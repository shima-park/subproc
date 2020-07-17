package subproc

import (
	"math"
	"os"
	"time"

	"github.com/google/uuid"
)

type SubProc interface {
	ID() string
	Cmd() string
	Options() CmdOptions
	Run()
	Stop()
	Restarts() int
	Status() SubProcStatus
	Error() error
}

type SubProcStatus string

const (
	SubProcStatusCreating         = "Creating"
	SubProcStatusRunning          = "Running"
	SubProcStatusCrashLoopBackOff = "CrashLoopBackOff"
	SubProcStatusExited           = "Exited"
)

type subproc struct {
	id       string
	cmd      string
	options  []CmdOption
	restarts int
	status   SubProcStatus
	error    error
	done     chan struct{}
}

func NewSubProc(cmd string, options ...CmdOption) SubProc {
	w := &subproc{
		id:      cmd + "-" + uuid.New().String(),
		cmd:     cmd,
		options: options,
		status:  SubProcStatusCreating,
		done:    make(chan struct{}),
	}

	return w
}

func (w *subproc) run() error {
	if w.isStopped() {
		return nil
	}

	cmd := NewCmd(w.cmd, w.options...)

	if err := cmd.Start(); err != nil {
		return err
	}

	closeSignal := make(chan struct{})
	defer close(closeSignal)
	go func() {
		select {
		case <-w.done:
			cmd.Process.Signal(os.Interrupt)
		case <-closeSignal:
			return
		}
	}()

	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func (w *subproc) Run() {
	for ; w.restarts < int(math.MaxInt32); w.restarts++ {
		w.error = nil
		w.status = SubProcStatusRunning
		err := w.run()
		w.status = SubProcStatusExited
		if err == nil || w.isStopped() {
			return
		}
		w.error = err
		w.status = SubProcStatusCrashLoopBackOff

		waitTime := time.Duration(math.Pow(2, float64(w.restarts))) * 100 * time.Millisecond
		select {
		case <-w.done:
			return
		case <-time.After(waitTime):

		}
	}
	return
}

func (w *subproc) Stop() {
	if w.isStopped() {
		return
	}

	close(w.done)
}

func (w *subproc) isStopped() bool {
	select {
	case <-w.done:
		return true
	default:
	}
	return false
}

func (w *subproc) ID() string {
	return w.id
}

func (w *subproc) Options() CmdOptions {
	var options CmdOptions
	for _, opt := range w.options {
		opt(&options)
	}
	return options
}

func (w *subproc) Cmd() string {
	return w.cmd
}

func (w *subproc) Status() SubProcStatus {
	return w.status
}

func (w *subproc) Restarts() int {
	return w.restarts
}

func (w *subproc) Error() error {
	return w.error
}
