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
	Containers map[string]*docker.Container
	Image      string
	Capacity   int
	Standby    int
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

func NewPool(config *Config, client *docker.Client, image string, capacity int, standby int) (*Pool, error) {
	_, err := findImage(client, image)
	if err != nil {
		return nil, err
	}

	if standby <= 60 {
		standby = 86400
	}

	pool := &Pool{
		Config:     config,
		Client:     client,
		Containers: map[string]*docker.Container{},
		Image:      image,
		Capacity:   capacity,
		Standby:    standby,
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
	container, err := CreateContainer(pool.Client, pool.Config, pool.Image, pool.Standby, "")
	if err != nil {
		return err
	}

	if err = pool.Client.StartContainer(container.ID, nil); err != nil {
		return err
	}

	pool.Lock()
	defer pool.Unlock()

	pool.Containers[container.ID] = container
	return nil
}

func (pool *Pool) Fill() {
	num := pool.Capacity - len(pool.Containers)

	// Pool is full
	if num <= 0 {
		return
	}

	log.Printf("adding %v containers to %v pool, standby: %vs\n", num, pool.Image, pool.Standby)

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
		go destroyContainer(pool.Client, id)
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

			if event.Status == "die" {
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
		if cfg.Capacity < 1 {
			continue
		}

		log.Println("initializing pool for:", cfg.Image)

		pool, err := NewPool(config, client, cfg.Image, cfg.Capacity, cfg.Standby)
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
