package main

import (
	"fmt"

	"os"
	"time"

	"github.com/shima-park/subproc"
)

func init() {
	subproc.Register("hello", hello)
	if subproc.Init() {
		os.Exit(0)
	}
}

func hello() {
	for {
		fmt.Println("hello world", time.Now())
		time.Sleep(time.Second)
	}
}

func main() {
	m := subproc.NewSubProcManager()
	m.ParallelismRun(3, "hello")

	for cmd, procs := range m.List() {
		for _, p := range procs {
			fmt.Println("cmd:", cmd, "id:", p.ID(), "status:", p.Status())
		}
	}

	select {}
}
