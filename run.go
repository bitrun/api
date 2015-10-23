package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

type Run struct {
	Id         string
	VolumePath string
	Config     *Config
	Container  *docker.Container
	Client     *docker.Client
	Request    *Request
	Done       chan bool
}

type RunResult struct {
	ExitCode int    `json:"exit_code"`
	Output   []byte `json:"output"`
	Duration string `json:"-"`
}

type Done struct {
	*RunResult
	error
}

func NewRun(config *Config, client *docker.Client, req *Request) *Run {
	id, _ := randomHex(20)

	return &Run{
		Id:         id,
		Config:     config,
		Client:     client,
		VolumePath: fmt.Sprintf("%s/%s", config.SharedPath, id),
		Request:    req,
		Done:       make(chan bool),
	}
}

func (run *Run) Setup() error {
	fullPath := fmt.Sprintf("%s/%s", run.VolumePath, run.Request.Filename)

	if err := os.Mkdir(run.VolumePath, 0777); err != nil {
		return err
	}

	if err := ioutil.WriteFile(fullPath, []byte(run.Request.Content), 0666); err != nil {
		return err
	}

	opts := docker.CreateContainerOptions{
		HostConfig: &docker.HostConfig{
			Binds: []string{
				run.VolumePath + ":/code",
				run.VolumePath + ":/tmp",
			},
			ReadonlyRootfs: true,
			Memory:         run.Config.MemoryLimit,
			MemorySwap:     0,
		},
		Config: &docker.Config{
			Hostname:        "bitrun",
			Image:           run.Request.Image,
			AttachStdout:    true,
			AttachStderr:    true,
			AttachStdin:     false,
			OpenStdin:       false,
			Tty:             true,
			NetworkDisabled: run.Config.NetworkDisabled,
			WorkingDir:      "/code",
			Cmd:             []string{"bash", "-c", run.Request.Command},
		},
	}

	container, err := run.Client.CreateContainer(opts)
	if err != nil {
		return err
	}

	run.Container = container
	return nil
}

func (run *Run) Start() (*RunResult, error) {
	ts := time.Now()

	err := run.Client.StartContainer(run.Container.ID, run.Container.HostConfig)
	if err != nil {
		return nil, err
	}

	result := RunResult{}

	exitCode, err := run.Client.WaitContainer(run.Container.ID)
	if err != nil {
		return nil, err
	}

	result.Duration = time.Now().Sub(ts).String()
	result.ExitCode = exitCode

	buff := bytes.NewBuffer([]byte{})

	err = run.Client.Logs(docker.LogsOptions{
		Container:    run.Container.ID,
		Stdout:       true,
		Stderr:       true,
		OutputStream: buff,
		ErrorStream:  buff,
		RawTerminal:  true,
	})

	if err != nil {
		return nil, err
	}

	result.Output = buff.Bytes()
	return &result, nil
}

func (run *Run) StartWithTimeout(duration time.Duration) (*RunResult, error) {
	timeout := time.After(duration)
	chDone := make(chan Done)

	go func() {
		res, err := run.Start()
		chDone <- Done{res, err}
	}()

	select {
	case done := <-chDone:
		return done.RunResult, done.error
	case <-timeout:
		return nil, fmt.Errorf("Operation timed out after %s", duration.String())
	}
}

func (run *Run) Destroy() error {
	if run.Container != nil {
		run.Client.RemoveContainer(docker.RemoveContainerOptions{
			ID:            run.Container.ID,
			RemoveVolumes: true,
			Force:         true,
		})
	}

	return os.RemoveAll(run.VolumePath)
}
