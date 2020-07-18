package subproc

import (
	"encoding/json"
	"os"
)

type CmdResult struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

func getResult(pipeReader *os.File, resp interface{}) CmdResult {
	var r CmdResult
	r.Data = resp
	if err := json.NewDecoder(pipeReader).Decode(&r); err != nil {
		r.Error = err.Error()
	}

	return r
}

func SuccCmd(data interface{}) {
	respondCmd(0, CmdResult{Data: data})
}

func FailCmd(err error) {
	if err != nil {
		respondCmd(1, CmdResult{Error: err.Error()})
	}
}

func respondCmd(code int, r CmdResult) {
	// 子进程默认继承stdin, stdout, stderr
	// 所以其他的ExtraFile的索引从3开始
	pipe := os.NewFile(3, "pipe")
	if err := json.NewEncoder(pipe).Encode(r); err != nil {
		panic(err)
	}

	os.Exit(code)
}
