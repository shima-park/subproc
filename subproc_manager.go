package subproc

import (
	"errors"
	"sync"
)

var ErrSubProcManagerIsClosed = errors.New("SubProcManager is closed")

type SubProcManager interface {
	Run(cmd string, opts ...CmdOption) error
	Kill(...MatchOption) error
	List(...MatchOption) map[string][]SubProc
	Restart(...MatchOption) error
	Wait()
	Stop()
}

type handler func(cmd string, i int)

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
		return ErrSubProcManagerIsClosed
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	m.subprocs[cmd] = append(m.subprocs[cmd], m.newSubProc(cmd, opts...))
	return nil
}

func (m *subProcManager) newSubProc(cmd string, opts ...CmdOption) SubProc {
	var options CmdOptions
	for _, opt := range opts {
		opt(&options)
	}

	opts = append(
		opts,
		StartBefore(func() error {
			m.wg.Add(1)
			if options.StartBefore != nil {
				return options.StartBefore()
			}
			return nil
		}),
		FinalHook(func(err error) {
			defer m.wg.Done()
			if options.FinalHook != nil {
				options.FinalHook(err)
			}
		}),
	)

	sp := NewSubProc(cmd, opts...)

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		sp.Run()
	}()

	return sp
}

func (m *subProcManager) iteratorCheckIsClosed(isReadlock bool, handle handler, opts ...MatchOption) error {
	if m.isStopped() {
		return ErrSubProcManagerIsClosed
	}

	return m.iterator(isReadlock, handle, opts...)
}

func (m *subProcManager) iterator(isReadlock bool, handle handler, opts ...MatchOption) error {
	if isReadlock {
		m.lock.RLock()
		defer m.lock.RUnlock()
	} else {
		m.lock.Lock()
		defer m.lock.Unlock()
	}

	options := NewMatchOptions(opts...)

	var matched bool
	for cmd := range m.subprocs {
		for i := len(m.subprocs[cmd]) - 1; i >= 0; i-- {
			subproc := m.subprocs[cmd][i]
			if options.Match(subproc) {
				if !matched {
					matched = true
				}
				handle(cmd, i)
			}
		}
	}
	if !matched {
		return errors.New("no such sub process")
	}
	return nil
}

func (m *subProcManager) Kill(opts ...MatchOption) error {
	return m.iteratorCheckIsClosed(false, m.kill, opts...)
}

func (m *subProcManager) kill(cmd string, i int) {
	m.subprocs[cmd][i].Stop()
	m.subprocs[cmd] = append(m.subprocs[cmd][:i], m.subprocs[cmd][i+1:]...)
	if len(m.subprocs[cmd]) == 0 {
		delete(m.subprocs, cmd)
	}
}

func (m *subProcManager) List(opts ...MatchOption) map[string][]SubProc {
	list := map[string][]SubProc{}
	_ = m.iteratorCheckIsClosed(true, func(cmd string, i int) {
		list[cmd] = append(list[cmd], m.subprocs[cmd][i])
	}, opts...)
	return list
}

func (m *subProcManager) Restart(opts ...MatchOption) error {
	return m.iteratorCheckIsClosed(false, m.restart, opts...)
}

func (m *subProcManager) restart(cmd string, i int) {
	subproc := m.subprocs[cmd][i]
	subproc.Stop()
	m.subprocs[cmd][i] = m.newSubProc(subproc.Cmd(), subproc.Options()...)
}

func (m *subProcManager) Wait() {
	m.wg.Wait()
}

func (m *subProcManager) Stop() {
	if m.isStopped() {
		return
	}

	close(m.done)

	_ = m.iterator(false, m.kill, WithMatchAll())

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
