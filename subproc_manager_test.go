package subproc

import (
	"fmt"
	"os"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func init() {
	Register("hello", hello)
	Register("hello2", hello)
	if Init() {
		os.Exit(0)
	}
}

func hello() {
	for {
		fmt.Println("hello world", time.Now())
		time.Sleep(time.Second)
	}
}

func TestSubProcManager(t *testing.T) {
	m := NewSubProcManager()

	m.Run("hello")
	assert.Equal(t, len(m.List()), 1)
	assert.Equal(t, len(m.List()["hello"]), 1)

	m.ParallelismRun(3, "hello")
	assert.Equal(t, len(m.List()), 1)
	assert.Equal(t, len(m.List()["hello"]), 3)

	m.ParallelismRun(2, "hello")
	assert.Equal(t, len(m.List()), 1)
	assert.Equal(t, len(m.List()["hello"]), 2)

	m.Kill(m.List()["hello"][0].ID())
	assert.Equal(t, len(m.List()), 1)
	assert.Equal(t, len(m.List()["hello"]), 1)

	m.ParallelismRun(1, "hello2")
	assert.Equal(t, len(m.List()), 2)
	assert.Equal(t, len(m.List()["hello2"]), 1)

	m.KillCmd("hello2")
	assert.Equal(t, len(m.List()), 1)

	m.Killall()
	assert.Equal(t, len(m.List()), 0)

	m.ParallelismRun(1, "hello")
	assert.Equal(t, len(m.List()), 1)
	assert.Equal(t, len(m.List()["hello"]), 1)

	m.Stop()
	assert.Equal(t, len(m.List()), 0)

	os.Exit(0)
}
