package main

import (
	"strconv"

	docker "github.com/fsouza/go-dockerclient"
	gin "github.com/gin-gonic/gin"
)

func errorReponse(err error, c *gin.Context) {
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
		errorReponse(err, c)
		return
	}

	config, err := c.Get("config")
	if err != nil {
		errorReponse(err, c)
		return
	}

	client, err := c.Get("client")
	if err != nil {
		errorReponse(err, c)
		return
	}

	run := NewRun(config.(*Config), client.(*docker.Client), req)
	defer run.Destroy()

	if err := run.Setup(); err != nil {
		errorReponse(err, c)
		return
	}

	result, err := run.Start()
	if err != nil {
		errorReponse(err, c)
		return
	}

	c.Writer.Header().Set("X-Run-ExitCode", strconv.Itoa(result.ExitCode))
	c.Writer.Header().Set("X-Run-Duration", result.Duration)

	c.String(200, result.Output)
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
