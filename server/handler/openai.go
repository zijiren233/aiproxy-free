package handler

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/labring/aiproxy-free/config"
	"github.com/labring/aiproxy-free/db"
	"github.com/labring/aiproxy-free/server/middleware"
	"github.com/labring/aiproxy-free/server/module"
	log "github.com/sirupsen/logrus"
)

const (
	CompletionsEndpoint = "/v1/chat/completions"
)

func ChatCompletionsHandler(c *gin.Context) {
	proxyToOpenAI(c)
}

func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Service is healthy",
	})
}

func UsageHandler(c *gin.Context) {
	namespace := c.GetString(middleware.NamespaceKey)
	if namespace == "" {
		c.JSON(http.StatusInternalServerError, module.NewInternalServerError())
		return
	}

	usedToday, nextResetTime, err := db.GetUsageInfo(namespace)
	if err != nil {
		log.Errorf("Failed to get usage info for namespace %s: %v", namespace, err)
		c.JSON(http.StatusInternalServerError, module.NewInternalServerError())
		return
	}

	totalLimit := config.DailyRequestLimit

	remainingToday := totalLimit - usedToday
	if remainingToday < 0 {
		remainingToday = 0
	}

	response := &module.UsageResponse{
		TotalLimit:     totalLimit,
		UsedToday:      usedToday,
		RemainingToday: remainingToday,
		NextResetTime:  nextResetTime.UnixMilli(),
	}

	c.JSON(http.StatusOK, response)
}

var proxyResponseHeaders = []string{"Content-Type", "Content-Length"}

func proxyToOpenAI(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Errorf("Failed to read request body: %v", err)
		c.JSON(http.StatusInternalServerError, module.NewInternalServerError())
		return
	}

	req, err := http.NewRequestWithContext(
		c.Request.Context(),
		c.Request.Method,
		config.UpstreamBaseURL+CompletionsEndpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		log.Errorf("Failed to create proxy request: %v", err)
		c.JSON(http.StatusInternalServerError, module.NewInternalServerError())
		return
	}

	req.Header.Set("Authorization", "Bearer "+config.UpstreamAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Failed to proxy request to upstream: %v", err)
		c.JSON(
			http.StatusBadGateway,
			module.NewBadGatewayError("Failed to connect to upstream API"),
		)

		return
	}
	defer resp.Body.Close()

	for _, h := range proxyResponseHeaders {
		c.Header(h, resp.Header.Get(h))
	}

	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
}
