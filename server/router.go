package server

import (
	"github.com/gin-gonic/gin"
	"github.com/labring/aiproxy-free/server/handler"
	"github.com/labring/aiproxy-free/server/middleware"
)

func SetRouter(router *gin.Engine) {
	api := router.Group("/api")
	{
		api.GET("/healthz", handler.HealthHandler)
	}

	v1 := router.Group("/v1")
	v1.Use(middleware.AuthMiddleware())
	v1.Use(middleware.RateLimitMiddleware())
	{
		v1.POST("/chat/completions", handler.ChatCompletionsHandler)
	}

	usage := router.Group("/usage")
	usage.Use(middleware.AuthMiddleware())
	{
		usage.GET("", handler.UsageHandler)
	}
}
