package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/labring/aiproxy-free/config"
	"github.com/labring/aiproxy-free/db"
	"github.com/labring/aiproxy-free/server/module"
	log "github.com/sirupsen/logrus"
)


func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.GetString(NamespaceKey)
		if namespace == "" {
			c.JSON(http.StatusInternalServerError, module.NewInternalServerError())
			c.Abort()
			return
		}

		if !checkRateLimit(namespace) {
			c.JSON(http.StatusTooManyRequests, module.NewRateLimitError(fmt.Sprintf("Daily request limit (%d) exceeded", config.DailyRequestLimit)))
			c.Abort()
			return
		}

		recordID, err := db.AddRequest(namespace)
		if err != nil {
			log.Errorf("Failed to record request: %v", err)
			c.JSON(http.StatusInternalServerError, module.NewInternalServerError())
			c.Abort()
			return
		}

		c.Next()

		// 如果响应状态不是200，删除之前添加的记录
		if c.Writer.Status() != http.StatusOK {
			if deleteErr := db.DeleteRequestByID(recordID); deleteErr != nil {
				log.Errorf("Failed to delete request record for failed response: %v", deleteErr)
			}
		}
	}
}

func checkRateLimit(namespace string) bool {
	count, err := db.CountRequestsToday(namespace)
	if err != nil {
		log.Errorf("Failed to check rate limit: %v", err)
		return false
	}
	return count < config.DailyRequestLimit
}
