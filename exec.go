package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

func (run *Run) StartExec(container *docker.Container) (*RunResult, error) {
	run.Container = container
	run.VolumePath = fmt.Sprintf("%s/%s", run.Config.SharedPath, container.Config.Labels["id"])
	fullPath := fmt.Sprintf("%s/%s", run.VolumePath, run.Request.Filename)

	if err := ioutil.WriteFile(fullPath, []byte(run.Request.Content), 0666); err != nil {
		return nil, err
	}

	ts := time.Now()

	exec, err := run.Client.CreateExec(docker.CreateExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"bash", "-c", run.Request.Command},
		Container:    container.ID,
	})

	if err != nil {
		return nil, err
	}

	buff := bytes.NewBuffer([]byte{})

	err = run.Client.StartExec(exec.ID, docker.StartExecOptions{
		Tty:          true,
		OutputStream: buff,
		ErrorStream:  buff,
		RawTerminal:  true,
	})

	if err != nil {
		return nil, err
	}

	result := RunResult{}

	result.Duration = time.Now().Sub(ts).String()
	result.Output = buff.Bytes()

	return &result, nil
}
