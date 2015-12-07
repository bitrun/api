package main

import (
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
	container, err := CreateContainer(run.Client, run.Config, run.Request.Image, 60)
	if err != nil {
		return err
	}

	volumePath := fmt.Sprintf("%s/%s", run.Config.SharedPath, container.Config.Labels["id"])
	fullPath := fmt.Sprintf("%s/%s", volumePath, run.Request.Filename)

	if err := ioutil.WriteFile(fullPath, []byte(run.Request.Content), 0666); err != nil {
		return err
	}

	run.Container = container

	if err := run.Client.StartContainer(container.ID, nil); err != nil {
		return err
	}

	return nil
}

func (run *Run) Start() (*RunResult, error) {
	return run.StartExec(run.Container)
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
		destroyContainer(run.Client, run.Container.ID)
	}

	return os.RemoveAll(run.VolumePath)
}

func destroyContainer(client *docker.Client, id string) error {
	return client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            id,
		RemoveVolumes: true,
		Force:         true,
	})
}
