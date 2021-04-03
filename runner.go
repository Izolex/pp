package main

import (
	"errors"
	"io"
	"io/fs"
	"os/exec"
	"sync"
)

type RunnerState int

const (
	RunnerStateRunning RunnerState = iota
	RunnerStateOK
	RunnerStateError
	RunnerStateWarning
)

type Runner struct {
	sync.RWMutex
	Broadcast *Broadcast
	Addr      string
	state     RunnerState
}

func NewRunner(addr string) *Runner {
	return &Runner{
		Broadcast: NewBroadcast(),
		Addr:      addr,
		state:     RunnerStateRunning,
	}
}

func (r *Runner) GetState() RunnerState {
	r.RLock()
	defer r.RUnlock()

	return r.state
}

func (r *Runner) Exec(cmd *exec.Cmd, codes *Codes) {
	defer r.Broadcast.Stop()

	go r.Broadcast.Start()
	go r.catchPipe(getPipe(cmd.StderrPipe()))
	go r.catchPipe(getPipe(cmd.StdoutPipe()))

	var err error
	var code int

	err = cmd.Start()
	if err != nil {
		r.Broadcast.Publish([]byte(err.Error()))
		r.setState(RunnerStateError)
	}

	err = cmd.Wait()
	if err == nil {
		r.Broadcast.Publish([]byte("exit status 0"))
		code = 0
	} else {
		r.Broadcast.Publish([]byte(err.Error()))

		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			r.Broadcast.Publish([]byte("pp: can not recognize status code :/"))
			code = -1
		}
	}

	switch true {
	case contains(code, codes.ok):
		r.setState(RunnerStateOK)
	case contains(code, codes.warning):
		r.setState(RunnerStateWarning)
	default:
		r.setState(RunnerStateError)
	}
}

func (r *Runner) catchPipe(reader io.ReadCloser) {
	var err error

	for {
		data := make([]byte, 1024)
		_, err = reader.Read(data)
		if err != nil {
			if err == io.EOF || err == io.ErrClosedPipe || errors.Is(err, fs.ErrClosed) {
				return
			}

			panic(err)
		}

		r.Broadcast.Publish(data)
	}
}

func (r *Runner) setState(state RunnerState) {
	r.Lock()
	defer r.Unlock()

	r.state = state
}

func getPipe(pipe io.ReadCloser, err error) io.ReadCloser {
	if err != nil {
		panic(err)
	}
	return pipe
}

func contains(code int, codes []int) bool {
	for _, v := range codes {
		if code == v {
			return true
		}
	}

	return false
}
