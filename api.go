package main

import (
	"strconv"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	gin "github.com/gin-gonic/gin"
)

func errorResponse(err error, c *gin.Context) {
	result := map[string]string{"error": err.Error()}
	c.JSON(400, result)
}

func setCorsHeaders(c *gin.Context) {
	c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
}

func HandleRun(c *gin.Context) {
	// We only need CORS for this endpoint
	setCorsHeaders(c)

	req, err := ParseRequest(c.Request)
	if err != nil {
		errorResponse(err, c)
		return
	}

	config, err := c.Get("config")
	if err != nil {
		errorResponse(err, c)
		return
	}

	client, err := c.Get("client")
	if err != nil {
		errorResponse(err, c)
		return
	}

	run := NewRun(config.(*Config), client.(*docker.Client), req)
	defer run.Destroy()

	if err := run.Setup(); err != nil {
		errorResponse(err, c)
		return
	}

	// TODO: make timeout configurable
	result, err := run.StartWithTimeout(time.Second * 10)
	if err != nil {
		errorResponse(err, c)
		return
	}

	c.Writer.Header().Set("X-Run-ExitCode", strconv.Itoa(result.ExitCode))
	c.Writer.Header().Set("X-Run-Duration", result.Duration)

	c.Data(200, req.Format, result.Output)
}

func RunApi(config *Config, client *docker.Client) {
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("config", config)
		c.Set("client", client)
	})

	router.POST("/run", HandleRun)
	router.Run("127.0.0.1:5000")
}
