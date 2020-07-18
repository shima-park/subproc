package subproc

import (
	"fmt"

	"github.com/docker/docker/pkg/reexec"
)

var defaultSubProcManager = NewSubProcManager()

func Register(name string, initializer func()) {
	reexec.Register(name, func() {
		defer func() {
			if r := recover(); r != nil {
				FailCmd(fmt.Errorf("%s", r))
			}
		}()
		initializer()
	})
}

func Init() bool {
	return reexec.Init()
}

func Run(cmd string, opts ...CmdOption) error {
	return defaultSubProcManager.Run(cmd, opts...)
}

func Kill(opts ...MatchOption) error {
	return defaultSubProcManager.Kill(opts...)
}

func Killall() error {
	return defaultSubProcManager.Kill(WithMatchAll())
}

func List(opts ...MatchOption) map[string][]SubProc {
	return defaultSubProcManager.List(opts...)
}

func ListAll() map[string][]SubProc {
	return defaultSubProcManager.List(WithMatchAll())
}

func Restart(opts ...MatchOption) error {
	return defaultSubProcManager.Restart(opts...)
}

func RestartAll() error {
	return defaultSubProcManager.Restart(WithMatchAll())
}

func Wait() {
	defaultSubProcManager.Wait()
}

func Stop() {
	defaultSubProcManager.Stop()
}
