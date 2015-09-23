package main

import (
	"fmt"
	"os"
)

type Config struct {
	DockerHost string
	SharedPath string
}

func NewConfig() (*Config, error) {
	cfg := Config{
		DockerHost: os.Getenv("DOCKER_HOST"),
		SharedPath: os.Getenv("SHARED_PATH"),
	}

	if cfg.DockerHost == "" {
		return nil, fmt.Errorf("Please set DOCKER_HOST environment variable")
	}

	if cfg.SharedPath == "" {
		return nil, fmt.Errorf("Please set SHARED_PATH environment variable")
	}

	cfg.SharedPath = expandPath(cfg.SharedPath)

	return &cfg, nil
}
