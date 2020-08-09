package subproc

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type SubProc interface {
	ID() string
	Cmd() string
	Options() []CmdOption
	Run()
	Stop()
	Metrics() SubProcMetrics
	Status() SubProcStatus
	Error() error
}

type SubProcMetrics struct {
	CreateTime time.Time
	StartTime  time.Time
	ExitTime   time.Time
	UpTime     time.Duration
	Restarts   int
}

type SubProcStatus string

const (
	SubProcStatusCreating         = "Creating"
	SubProcStatusRunning          = "Running"
	SubProcStatusCrashLoopBackOff = "CrashLoopBackOff"
	SubProcStatusExited           = "Exited"
)

type subproc struct {
	id      string
	cmd     string
	options []CmdOption
	metrics SubProcMetrics
	status  SubProcStatus
	error   error
	done    chan struct{}
}

func NewSubProc(cmd string, options ...CmdOption) SubProc {
	w := &subproc{
		id:      cmd + "-" + uuid.New().String(),
		cmd:     cmd,
		options: options,
		metrics: SubProcMetrics{
			CreateTime: time.Now(),
		},
		status: SubProcStatusCreating,
		done:   make(chan struct{}),
	}

	return w
}

func (w *subproc) Run() {
	for ; w.metrics.Restarts < int(math.MaxInt32); w.metrics.Restarts++ {
		w.error = nil
		w.status = SubProcStatusRunning
		err := w.run()
		w.status = SubProcStatusExited
		if err == nil || w.isStopped() {
			return
		}
		w.error = err
		w.status = SubProcStatusCrashLoopBackOff

		waitTime := time.Duration(math.Pow(2, float64(w.metrics.Restarts))) * 100 * time.Millisecond
		select {
		case <-w.done:
			return
		case <-time.After(waitTime):

		}
	}
}

func (w *subproc) run() error {
	if w.isStopped() {
		return nil
	}

	cmd := NewCmd(w.cmd, w.options...)

	w.metrics.StartTime = time.Now()
	if err := cmd.Start(); err != nil {
		return err
	}

	closeSignal := make(chan struct{})
	defer close(closeSignal)
	go func() {
		select {
		case <-w.done:
			_ = cmd.Stop()
		case <-closeSignal:
			return
		}
	}()

	err := cmd.Wait()
	w.metrics.ExitTime = time.Now()
	if err != nil {
		return err
	}
	return nil
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

func (w *subproc) Options() []CmdOption {
	return w.options
}

func (w *subproc) Cmd() string {
	return w.cmd
}

func (w *subproc) Status() SubProcStatus {
	return w.status
}

func (w *subproc) Metrics() SubProcMetrics {
	if !w.metrics.ExitTime.IsZero() {
		w.metrics.UpTime = w.metrics.ExitTime.Sub(w.metrics.StartTime)
	} else {
		w.metrics.UpTime = time.Since(w.metrics.StartTime)
	}
	return w.metrics
}

func (w *subproc) Error() error {
	return w.error
}
