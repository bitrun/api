package main

import (
	"fmt"
	"log"
	"os"

	docker "github.com/fsouza/go-dockerclient"
)

const VERSION = "0.1.0"

func requireEnvVar(name string) {
	if os.Getenv(name) == "" {
		err := fmt.Errorf("Please set %s environment variable", name)
		log.Fatalln(err)
	}
}

func getConfig() *Config {
	if os.Getenv("CONFIG") != "" {
		config, err := NewConfigFromFile(os.Getenv("CONFIG"))
		if err != nil {
			log.Fatalln(err)
		}

		return config
	}

	requireEnvVar("DOCKER_HOST")
	requireEnvVar("SHARED_PATH")
	return NewConfig()
}

func checkImages(client *docker.Client) error {
	images, err := client.ListImages(docker.ListImagesOptions{})
	if err != nil {
		log.Fatalln(err)
	}

	imagesWithTags := map[string]bool{}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			imagesWithTags[tag] = true
		}
	}

	fmt.Println("checking images...")
	for _, lang := range Extensions {
		if imagesWithTags[lang.Image] == true {
			log.Printf("image %s exists", lang.Image)
		} else {
			return fmt.Errorf("image %s does not exist", lang.Image)
		}
	}

	return nil
}

func main() {
	log.Printf("bitrun api v%s\n", VERSION)

	err := LoadLanguages("./languages.json")
	if err != nil {
		log.Fatalln(err)
	}

	config := getConfig()

	client, err := docker.NewClient(config.DockerHost)
	if err != nil {
		log.Fatalln(err)
	}

	err = checkImages(client)
	if err != nil {
		log.Fatalln(err)
	}

	go RunPool(config, client)
	RunApi(config, client)
}
