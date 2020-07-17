package subproc

import (
	"github.com/docker/docker/pkg/reexec"
)

var defaultSubProcManager = NewSubProcManager()

func Register(name string, initializer func()) {
	reexec.Register(name, initializer)
}

func Init() bool {
	return reexec.Init()
}

func Run(cmd string, opts ...CmdOption) error {
	return defaultSubProcManager.Run(cmd, opts...)
}

func ParallelismRun(parallelism int, cmd string, opts ...CmdOption) error {
	return defaultSubProcManager.ParallelismRun(parallelism, cmd, opts...)
}

func Kill(ids ...string) {
	defaultSubProcManager.Kill(ids...)
}

func KillCmd(cmds ...string) {
	defaultSubProcManager.KillCmd(cmds...)
}

func Killall() {
	defaultSubProcManager.Killall()
}

func List() map[string][]SubProc {
	return defaultSubProcManager.List()
}

func Stop() {
	defaultSubProcManager.Stop()
}
