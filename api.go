package main

import (
	"fmt"
	"strconv"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	gin "github.com/gin-gonic/gin"
)

func errorResponse(err error, c *gin.Context) {
	result := map[string]string{"error": err.Error()}
	c.JSON(400, result)
}

func performRun(run *Run) (*RunResult, error) {
	// Try to get a warmed-up container for the run
	if pools[run.Request.Image] != nil {
		container, err := pools[run.Request.Image].Get()

		if err == nil {
			result, err := run.StartExec(container)
			return result, err
		}
	}

	if err := run.Setup(); err != nil {
		return nil, err
	}

	return run.StartWithTimeout(run.Config.RunDuration)
}

func HandleRun(c *gin.Context) {
	req, err := ParseRequest(c.Request)
	if err != nil {
		errorResponse(err, c)
		return
	}

	config, exists := c.Get("config")
	if !exists {
		errorResponse(fmt.Errorf("Cant get config"), c)
		return
	}

	client, exists := c.Get("client")
	if !exists {
		errorResponse(fmt.Errorf("Cant get client"), c)
		return
	}

	run := NewRun(config.(*Config), client.(*docker.Client), req)
	defer run.Destroy()

	result, err := performRun(run)
	if err != nil {
		errorResponse(err, c)
		return
	}

	c.Header("X-Run-Command", req.Command)
	c.Header("X-Run-ExitCode", strconv.Itoa(result.ExitCode))
	c.Header("X-Run-Duration", result.Duration)

	c.Data(200, req.Format, result.Output)
}

func HandleConfig(c *gin.Context) {
	c.JSON(200, Extensions)
}

func throttleMiddleware(throttler *Throttler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := strings.Split(c.Request.RemoteAddr, ":")[0]

		if err := throttler.Add(ip); err != nil {
			errorResponse(err, c)
			c.Abort()
			return
		}

		c.Next()
		throttler.Remove(ip)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Expose-Headers", "*")
	}
}

func RunApi(config *Config, client *docker.Client) {
	throttler := NewThrottler(config.ThrottleConcurrency, config.ThrottleQuota)
	throttler.StartPeriodicFlush()

	router := gin.Default()

	v1 := router.Group("/api/v1/")
	{
		v1.Use(corsMiddleware())
		v1.Use(throttleMiddleware(throttler))

		v1.Use(func(c *gin.Context) {
			c.Set("config", config)
			c.Set("client", client)
		})

		v1.GET("/config", HandleConfig)
		v1.POST("/run", HandleRun)
	}

	router.Run("127.0.0.1:5000")
}
