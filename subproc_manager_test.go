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
	defer func() {
		fmt.Println("=====stop")
		m.Stop()
		assert.Equal(t, len(m.List()), 0)
	}()

	err := m.Run("hello")
	assert.NilError(t, err)
	assert.Equal(t, len(m.List(WithMatchAll())), 1)
	assert.Equal(t, len(m.List(WithMatchAll())["hello"]), 1)

	err = m.Run("hello")
	assert.NilError(t, err)
	assert.Equal(t, len(m.List(WithMatchAll())), 1)
	assert.Equal(t, len(m.List(WithMatchAll())["hello"]), 2)

	err = m.Kill(WithIDs(m.List(WithMatchAll())["hello"][0].ID()))
	assert.NilError(t, err)
	assert.Equal(t, len(m.List(WithMatchAll())), 1)
	assert.Equal(t, len(m.List(WithMatchAll())["hello"]), 1)

	err = m.Run("hello2")
	assert.NilError(t, err)
	assert.Equal(t, len(m.List(WithMatchAll())), 2)
	assert.Equal(t, len(m.List(WithMatchAll())["hello2"]), 1)

	err = m.Kill(WithCmds("hello2"))
	assert.NilError(t, err)
	assert.Equal(t, len(m.List(WithMatchAll())), 1)

	err = m.Kill(WithMatchAll())
	assert.NilError(t, err)
	assert.Equal(t, len(m.List(WithMatchAll())), 0)

	err = m.Run("hello")
	assert.NilError(t, err)
	assert.Equal(t, len(m.List(WithMatchAll())), 1)
	assert.Equal(t, len(m.List(WithMatchAll())["hello"]), 1)

	os.Exit(0)
}
