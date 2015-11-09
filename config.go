package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

type PoolConfig struct {
	Image    string `json:"image"`
	Capacity int    `json:"capacity"`
	Standby  int    `json:"standby"`
}

type Config struct {
	DockerHost          string        `json:"docker_host"`
	SharedPath          string        `json:"shared_path"`
	RunDuration         time.Duration `json:"run_duration"`
	ThrottleQuota       int           `json:"throttle_quota"`
	ThrottleConcurrency int           `json:"throttle_concurrency"`
	NetworkDisabled     bool          `json:"network_disabled"`
	MemoryLimit         int64         `json:"memory_limit"`
	Pools               []PoolConfig  `json:"pools"`
}

func NewConfig() *Config {
	cfg := Config{
		DockerHost: os.Getenv("DOCKER_HOST"),
		SharedPath: os.Getenv("SHARED_PATH"),
	}

	cfg.SharedPath = expandPath(cfg.SharedPath)
	cfg.RunDuration = time.Second * 10
	cfg.ThrottleQuota = 5
	cfg.ThrottleConcurrency = 1
	cfg.NetworkDisabled = true
	cfg.MemoryLimit = 67108864
	cfg.Pools = []PoolConfig{}

	return &cfg
}

func NewConfigFromFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := Config{}

	err = json.Unmarshal(data, &config)

	if err == nil {
		config.SharedPath = expandPath(config.SharedPath)
		config.RunDuration = config.RunDuration * time.Second
	}

	return &config, err
}
