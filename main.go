package main

import (
	"log"

	docker "github.com/fsouza/go-dockerclient"
)

const VERSION = "0.1.0"

func main() {
	err := LoadLanguages("./languages.json")
	if err != nil {
		log.Fatalln(err)
	}

	config, err := NewConfig()
	if err != nil {
		log.Fatalln(err)
	}

	client, err := docker.NewClient(config.DockerHost)
	if err != nil {
		log.Fatalln(err)
	}

	RunApi(config, client)
}
