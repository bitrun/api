package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
)

const VERSION = "0.3.1"

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

func pullImage(name string, client *docker.Client) error {
	chunks := strings.Split(name, ":")

	imgName := chunks[0]
	imgTag := "latest"

	if len(chunks) == 2 {
		imgTag = chunks[1]
	}

	auth := docker.AuthConfiguration{}
	opts := docker.PullImageOptions{
		Repository:   imgName,
		Tag:          imgTag,
		OutputStream: os.Stdout,
	}

	return client.PullImage(opts, auth)
}

func checkImages(client *docker.Client, config *Config) error {
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
			if config.FetchImages {
				log.Println("pulling", lang.Image, "image...")
				err := pullImage(lang.Image, client)
				if err != nil {
					log.Fatalln(err)
				}
			} else {
				return fmt.Errorf("image %s does not exist", lang.Image)
			}
		}
	}

	return nil
}

func main() {
	log.Printf("bitrun api v%s\n", VERSION)

	config := getConfig()

	err := LoadLanguages(config.LanguagesPath)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := docker.NewClient(config.DockerHost)
	if err != nil {
		log.Fatalln(err)
	}

	err = checkImages(client, config)
	if err != nil {
		log.Fatalln(err)
	}

	go RunPool(config, client)
	RunApi(config, client)
}
