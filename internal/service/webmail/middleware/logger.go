package middleware

import (
	"easymail/internal/easylog"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func Access(_log *easylog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// request start time
		startTime := time.Now()

		// after handle
		c.Next()

		// request end time
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// parse request struct
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if statusCode >= http.StatusBadRequest {
			_log.Warningf("ip=%s; method=%s; path=%s; status_code=%d; latency=%s; error=%s\n", clientIP, method, path, statusCode, latency.String(), errorMessage)
		} else {
			_log.Infof("ip=%s; method=%s; path=%s; status_code=%d; latency=%s; error=%s\n", clientIP, method, path, statusCode, latency.String(), errorMessage)
		}
	}
}
