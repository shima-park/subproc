package subproc

import (
	"errors"
	"sync"
)

type SubProcManager interface {
	Run(cmd string, opts ...CmdOption) error
	ParallelismRun(parallelism int, cmd string, opts ...CmdOption) error
	Kill(ids ...string)
	KillCmd(cmds ...string)
	Killall()
	List() map[string][]SubProc
	Stop()
}

type subProcManager struct {
	lock     sync.RWMutex
	subprocs map[string][]SubProc
	wg       *sync.WaitGroup
	done     chan struct{}
}

func NewSubProcManager() SubProcManager {
	return &subProcManager{
		subprocs: map[string][]SubProc{},
		done:     make(chan struct{}),
		wg:       &sync.WaitGroup{},
	}
}

func (m *subProcManager) Run(cmd string, opts ...CmdOption) error {
	if m.isStopped() {
		return errors.New("SubProcManager is closed")
	}

	m.lock.Lock()
	m.subprocs[cmd] = append(m.subprocs[cmd], m.newSubProc(cmd, opts...))
	m.lock.Unlock()

	return nil
}

func (m *subProcManager) newSubProc(cmd string, opts ...CmdOption) SubProc {
	var options CmdOptions
	for _, opt := range opts {
		opt(&options)
	}

	opts = append(
		opts,
		PreHook(func() error {
			m.wg.Add(1)
			if options.PreHook != nil {
				return options.PreHook()
			}
			return nil
		}),
		FinalHook(func(err error) {
			if options.FinalHook != nil {
				options.FinalHook(err)
			}
			m.wg.Done()
		}),
	)

	sp := NewSubProc(cmd, opts...)

	go sp.Run()

	return sp
}

func (m *subProcManager) ParallelismRun(parallelism int, cmd string, opts ...CmdOption) error {
	if m.isStopped() {
		return errors.New("SubProcManager is closed")
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	size := parallelism
	curSize := len(m.subprocs[cmd])

	if size > curSize { // 扩容
		delta := size - curSize

		for i := 0; i < delta; i++ {
			m.subprocs[cmd] = append(m.subprocs[cmd], m.newSubProc(cmd, opts...))
		}

	} else if size < curSize { // 缩容
		var removed []SubProc
		m.subprocs[cmd], removed = m.subprocs[cmd][:size], m.subprocs[cmd][size:]
		for _, w := range removed {
			w.Stop()
		}
	}

	// reload
	for i := 0; i < min(curSize, len(m.subprocs[cmd])); i++ {
		w := m.subprocs[cmd][i]
		w.Stop()
		m.subprocs[cmd][i] = m.newSubProc(cmd, opts...)
	}
	return nil
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func (m *subProcManager) kill(match func(SubProc) bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for cmd, _ := range m.subprocs {
		for i := len(m.subprocs[cmd]) - 1; i >= 0; i-- {
			subproc := m.subprocs[cmd][i]
			if match(subproc) {
				m.remove(cmd, i)
			}
		}
	}
}

func (m *subProcManager) remove(cmd string, i int) {
	m.subprocs[cmd][i].Stop()
	m.subprocs[cmd] = append(m.subprocs[cmd][:i], m.subprocs[cmd][i+1:]...)
	if len(m.subprocs[cmd]) == 0 {
		delete(m.subprocs, cmd)
	}
}

func (m *subProcManager) Kill(ids ...string) {
	m.kill(func(sp SubProc) bool {
		for _, id := range ids {
			if id == sp.ID() {
				return true
			}
		}
		return false
	})
}

func (m *subProcManager) KillCmd(cmds ...string) {
	m.kill(func(sp SubProc) bool {
		for _, cmd := range cmds {
			if cmd == sp.Cmd() {
				return true
			}
		}
		return false
	})
}

func (m *subProcManager) Killall() {
	m.kill(func(sp SubProc) bool {
		return true
	})
}

func (m *subProcManager) List() map[string][]SubProc {
	var list = map[string][]SubProc{}
	m.lock.RLock()
	for cmd, subprocs := range m.subprocs {
		list[cmd] = subprocs
	}
	m.lock.RUnlock()
	return list
}

func (m *subProcManager) Stop() {
	if m.isStopped() {
		return
	}

	close(m.done)

	m.Killall()

	m.wg.Wait()
}

func (m *subProcManager) isStopped() bool {
	select {
	case <-m.done:
		return true
	default:

	}
	return false
}
