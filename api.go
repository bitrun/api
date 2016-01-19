package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	gin "github.com/gin-gonic/gin"
)

func errorResponse(status int, err error, c *gin.Context) {
	result := map[string]string{"error": err.Error()}
	c.JSON(status, result)
}

func performRun(run *Run) (*RunResult, error) {
	// Try to get a warmed-up container for the run
	if pools[run.Request.Image] != nil {
		container, err := pools[run.Request.Image].Get()

		if err == nil {
			log.Println("got warmed-up container for image:", run.Request.Image, container.ID)
			result, err := run.StartExecWithTimeout(container)
			return result, err
		}
	}

	log.Println("setting up container for image:", run.Request.Image)
	if err := run.Setup(); err != nil {
		return nil, err
	}

	return run.StartWithTimeout()
}

func HandleRun(c *gin.Context) {
	req, err := ParseRequest(c.Request)
	if err != nil {
		errorResponse(400, err, c)
		return
	}

	config, exists := c.Get("config")
	if !exists {
		errorResponse(400, fmt.Errorf("Cant get config"), c)
		return
	}

	client, exists := c.Get("client")
	if !exists {
		errorResponse(400, fmt.Errorf("Cant get client"), c)
		return
	}

	run := NewRun(config.(*Config), client.(*docker.Client), req)
	defer run.Destroy()

	result, err := performRun(run)
	if err != nil {
		errorResponse(400, err, c)
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

func authMiddleware(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.ApiToken != "" {
			token := c.Request.FormValue("api_token")

			if token != config.ApiToken {
				errorResponse(400, fmt.Errorf("Api token is invalid"), c)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func throttleMiddleware(throttler *Throttler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := strings.Split(c.Request.RemoteAddr, ":")[0]

		if err := throttler.Add(ip); err != nil {
			errorResponse(429, err, c)
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

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	v1 := router.Group("/api/v1/")
	{
		v1.Use(authMiddleware(config))
		v1.Use(corsMiddleware())
		v1.Use(throttleMiddleware(throttler))

		v1.Use(func(c *gin.Context) {
			c.Set("config", config)
			c.Set("client", client)
		})

		v1.GET("/config", HandleConfig)
		v1.POST("/run", HandleRun)
	}

	fmt.Println("starting server on", config.Listen)
	router.Run(config.Listen)
}
