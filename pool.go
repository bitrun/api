package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

var pools map[string]*Pool

type Pool struct {
	Config     *Config
	Client     *docker.Client
	Containers map[string]*docker.Container
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
		Containers: map[string]*docker.Container{},
		Image:      image,
		Capacity:   capacity,
	}

	return pool, nil
}

func (pool *Pool) Exists(id string) bool {
	return pool.Containers[id] != nil
}

func (pool *Pool) Load() error {
	pool.Lock()
	defer pool.Unlock()

	containers, err := pool.Client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return err
	}

	for _, c := range containers {
		if c.Image == pool.Image && c.Labels["id"] != "" {
			pool.Containers[c.ID] = &docker.Container{
				ID: c.ID,
				Config: &docker.Config{
					Labels: c.Labels,
				},
			}
		}
	}

	return nil
}

func (pool *Pool) Add() error {
	id, _ := randomHex(20)
	volumePath := fmt.Sprintf("%s/%s", pool.Config.SharedPath, id)

	if err := os.Mkdir(volumePath, 0777); err != nil {
		return err
	}

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
			Labels:          map[string]string{"id": id},
			AttachStdout:    true,
			AttachStderr:    true,
			AttachStdin:     false,
			OpenStdin:       false,
			Tty:             true,
			NetworkDisabled: pool.Config.NetworkDisabled,
			WorkingDir:      "/code",
			Cmd:             []string{"sleep", "999999999"},
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

	// We need this!
	container.Config = opts.Config

	pool.Containers[container.ID] = container
	return nil
}

func (pool *Pool) Fill() {
	num := pool.Capacity - len(pool.Containers)

	// Pool is full
	if num <= 0 {
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

func (pool *Pool) Remove(id string) {
	pool.Lock()
	defer pool.Unlock()

	if pool.Containers[id] != nil {
		delete(pool.Containers, id)
	}
}

func (pool *Pool) Get() (*docker.Container, error) {
	pool.Lock()
	defer pool.Unlock()

	var container *docker.Container

	for _, v := range pool.Containers {
		container = v
		break
	}

	if container != nil {
		delete(pool.Containers, container.ID)
		return container, nil
	}

	return nil, fmt.Errorf("no contaienrs are available")
}

func RunPool(config *Config, client *docker.Client) {
	chEvents := make(chan *docker.APIEvents)
	pools = make(map[string]*Pool)

	// Setup docker event listener
	if err := client.AddEventListener(chEvents); err != nil {
		log.Fatalln(err)
	}

	go func() {
		for {
			event := <-chEvents
			if event == nil {
				continue
			}

			if event.Status == "destroy" {
				for _, pool := range pools {
					if pool.Exists(event.ID) {
						log.Println("pool's container got destroyed:", event.ID)
						pool.Remove(event.ID)
					}
				}
			}
		}
	}()

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
