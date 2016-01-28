package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

func CreateContainer(client *docker.Client, config *Config, image string, standby int, env string) (*docker.Container, error) {
	id, _ := randomHex(20)
	volumePath := fmt.Sprintf("%s/%s", config.SharedPath, id)
	name := fmt.Sprintf("bitrun-%v", time.Now().UnixNano())

	if err := os.Mkdir(volumePath, 0777); err != nil {
		return nil, err
	}

	opts := docker.CreateContainerOptions{
		Name: name,
		HostConfig: &docker.HostConfig{
			Binds: []string{
				volumePath + ":/code",
				volumePath + ":/tmp",
			},
			ReadonlyRootfs: true,
			Memory:         config.MemoryLimit,
			MemorySwap:     0,
		},
		Config: &docker.Config{
			Hostname:        "bitrun",
			Image:           image,
			Labels:          map[string]string{"id": id},
			AttachStdout:    false,
			AttachStderr:    false,
			AttachStdin:     false,
			Tty:             false,
			NetworkDisabled: config.NetworkDisabled,
			WorkingDir:      "/code",
			Cmd:             []string{"sleep", fmt.Sprintf("%v", standby)},
			Env:             strings.Split(env, "\n"),
		},
	}

	container, err := client.CreateContainer(opts)
	if err == nil {
		container.Config = opts.Config
	}

	return container, err
}
