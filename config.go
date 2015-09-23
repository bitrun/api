package main

import (
	"fmt"
	"os"
	"strings"
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

	// Expand home path if necessary
	if strings.Index(cfg.SharedPath, "~") == 0 {
		cfg.SharedPath = strings.Replace(cfg.SharedPath, "~", os.Getenv("HOME"), 1)
	}

	return &cfg, nil
}
