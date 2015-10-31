package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

var pools map[string]*Pool

type Pool struct {
	Config     *Config
	Client     *docker.Client
	Containers map[string]int
	Image      string
	Capacity   int
	sync.Mutex
}

func findImage(client *docker.Client, image string) (*docker.APIImages, error) {
	images, err := client.ListImages(docker.ListImagesOptions{})
	if err != nil {
		return nil, err
	}

	for _, img := range images {
		for _, t := range img.RepoTags {
			if t == image {
				return &img, nil
			}
		}
	}

	return nil, fmt.Errorf("invalid image:", image)
}

func NewPool(config *Config, client *docker.Client, image string, capacity int) (*Pool, error) {
	_, err := findImage(client, image)
	if err != nil {
		return nil, err
	}

	pool := &Pool{
		Config:     config,
		Client:     client,
		Containers: map[string]int{},
		Image:      image,
		Capacity:   capacity,
	}

	return pool, nil
}

func (pool *Pool) Load() error {
	pool.Lock()
	defer pool.Unlock()

	containers, err := pool.Client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return err
	}

	for _, c := range containers {
		if c.Image == pool.Image {
			pool.Containers[c.ID] = 0
		}
	}

	return nil
}

func (pool *Pool) Add() error {
	id, _ := randomHex(20)
	volumePath := fmt.Sprintf("%s/%s", pool.Config.SharedPath, id)

	opts := docker.CreateContainerOptions{
		Name: fmt.Sprintf("reserved-%v", time.Now().UnixNano()),
		HostConfig: &docker.HostConfig{
			Binds: []string{
				volumePath + ":/code",
				volumePath + ":/tmp",
			},
			ReadonlyRootfs: true,
			Memory:         pool.Config.MemoryLimit,
			MemorySwap:     0,
		},
		Config: &docker.Config{
			Hostname:        "bitrun",
			Image:           pool.Image,
			AttachStdout:    true,
			AttachStderr:    true,
			AttachStdin:     false,
			OpenStdin:       false,
			Tty:             true,
			NetworkDisabled: pool.Config.NetworkDisabled,
			WorkingDir:      "/code",
			Cmd:             []string{"sleep", "86400"},
		},
	}

	container, err := pool.Client.CreateContainer(opts)
	if err != nil {
		return err
	}

	err = pool.Client.StartContainer(container.ID, container.HostConfig)
	if err != nil {
		return err
	}

	pool.Lock()
	defer pool.Unlock()

	pool.Containers[container.ID] = 0
	return nil
}

func (pool *Pool) Fill() {
	num := pool.Capacity - len(pool.Containers)

	// Pool is full
	if num == 0 {
		return
	}

	log.Printf("adding %v containers to %v pool\n", num, pool.Image)

	for i := 0; i < num; i++ {
		err := pool.Add()
		if err != nil {
			log.Println("error while adding to pool:", err)
		}
	}
}

func (pool *Pool) Monitor() {
	for {
		pool.Fill()
		time.Sleep(time.Second * 3)
	}
}

func (pool *Pool) Get() (string, error) {
	pool.Lock()
	defer pool.Unlock()

	id := ""

	for k, v := range pool.Containers {
		if v == 0 {
			id = k
			break
		}
	}

	if id != "" {
		delete(pool.Containers, id)
		return id, nil
	}

	return "", fmt.Errorf("no contaienrs are available")
}

func RunPool(config *Config, client *docker.Client) {
	pools = make(map[string]*Pool)

	for _, cfg := range config.Pools {
		log.Println("initializing pool for:", cfg.Image)

		pool, err := NewPool(config, client, cfg.Image, cfg.Capacity)
		if err != nil {
			log.Fatalln(err)
		}

		err = pool.Load()
		if err != nil {
			log.Fatalln(err)
		}

		go pool.Monitor()
		pools[cfg.Image] = pool
	}
}
