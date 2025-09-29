package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/labring/aiproxy-free/utils"
	"github.com/sirupsen/logrus"
)

const RequestAt = "request_at"

func SetRequestAt(c *gin.Context, requestAt time.Time) {
	c.Set(RequestAt, requestAt)
}

func GetRequestAt(c *gin.Context) time.Time {
	return c.GetTime(RequestAt)
}

func NewLog(l *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		SetRequestAt(c, start)

		fields := utils.GetLogFields()
		defer func() {
			utils.PutLogFields(fields)
		}()

		entry := &logrus.Entry{
			Logger: l,
			Data:   fields,
		}
		utils.SetLogger(c.Request, entry)

		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		param := gin.LogFormatterParams{
			Request: c.Request,
			Keys:    c.Keys,
		}

		// Stop timer
		param.Latency = time.Since(start)

		param.ClientIP = c.ClientIP()
		param.Method = c.Request.Method
		param.StatusCode = c.Writer.Status()
		param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()

		param.BodySize = c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		param.Path = path

		logColor(entry, param)
	}
}

func logColor(log *logrus.Entry, p gin.LogFormatterParams) {
	str := formatter(p)

	code := p.StatusCode
	switch {
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		log.Error(str)
	default:
		log.Info(str)
	}
}

func formatter(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	return fmt.Sprintf("[GIN] |%s %3d %s| %10v | %15s |%s %-7s %s %#v\n%s",
		statusColor, param.StatusCode, resetColor,
		utils.TruncateDuration(param.Latency),
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}
