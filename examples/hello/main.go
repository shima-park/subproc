package main

import (
	"errors"
	"fmt"
	"strings"

	"os"
	"time"

	"github.com/shima-park/subproc"
)

func init() {
	subproc.Register("hello", hello)
	subproc.Register("mockPanic", mockPanic)
	subproc.Register("mockError", mockError)
	subproc.Register("mockSucc", mockSucc)
	if subproc.Init() {
		os.Exit(0)
	}
}

func hello() {
	for {
		fmt.Println(time.Now(), "hello world")
		time.Sleep(time.Second)
	}
}

func mockPanic() {
	panic("mock panic")
}

func mockError() {
	subproc.FailCmd(errors.New("mock error"))
}

func mockSucc() {
	subproc.SuccCmd("Hello world")
}

func printProcs(m subproc.SubProcManager) {
	go func() {
		for {
			fmt.Println(strings.Repeat("=", 20), "begin", strings.Repeat("=", 20))
			for cmd, procs := range m.List(subproc.WithMatchAll()) {
				for _, p := range procs {
					fmt.Println(
						"cmd:", cmd, "id:", p.ID(), "status:", p.Status(),
						"metrics:", fmt.Sprintf("%+v", p.Metrics()),
						"status:", p.Status(), "error:", p.Error(),
					)
				}
			}
			fmt.Println(strings.Repeat("=", 20), "end", strings.Repeat("=", 20))
			time.Sleep(time.Second)
		}
	}()
}

func main() {
	m := subproc.NewSubProcManager()
	defer m.Stop()

	m.Run("hello")     // Continuous running
	m.Run("mockPanic") // CrashLoopBackOff
	m.Run("mockError") // CrashLoopBackOff

	respCh := make(chan struct{})
	var resp string
	m.Run("mockSucc",
		subproc.FinalHook(func(error) {
			close(respCh)
		}),
		subproc.Response(&resp),
	) // Run once and exit
	<-respCh
	fmt.Println("resp: ", resp) // resp: Hello world

	printProcs(m)
	/* you will see
	cmd: hello      restart: 0  status: Running          error: <nil>
	cmd: mockSucc   restart: 0  status: Exited           error: <nil>
	cmd: mockPanic  restart: 3  status: CrashLoopBackOff error: Command Error: exit status 1, Message: mock panic
	cmd: mockError  restart: 3  status: CrashLoopBackOff error: Command Error: exit status 1, Message: mock error
	*/

	m.Wait()
}
