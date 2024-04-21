package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func Access() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// request start time
		startTime := time.Now()

		// after handle
		ctx.Next()

		// request end time
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// parse request struct
		clientIP := ctx.ClientIP()
		method := ctx.Request.Method
		path := ctx.Request.URL.Path
		statusCode := ctx.Writer.Status()
		errorMessage := ctx.Errors.ByType(gin.ErrorTypePrivate).String()

		// access log record by logrus
		entry := logrus.WithFields(logrus.Fields{
			"time":        endTime.Format("2006-01-02 15:04:05"),
			"ip":          clientIP,
			"method":      method,
			"path":        path,
			"status_code": statusCode,
			"latency":     latency.String(),
			"error":       errorMessage,
		})

		if statusCode >= http.StatusBadRequest {
			entry.Warn("HTTP request")
		} else {
			entry.Info("HTTP request")
		}
	}
}
