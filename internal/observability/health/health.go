package health

import (
	"context"
	"easymail/internal/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// LiveHandler indicates the process is running (no dependency checks).
func LiveHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// ReadyHandler indicates whether dependencies are available.
func ReadyHandler(c *gin.Context) {
	type dep struct {
		Status   string `json:"status"`
		Latency string `json:"latency"`
		Error   string `json:"error,omitempty"`
	}

	status := http.StatusOK
	deps := map[string]dep{}

	// DB
	start := time.Now()
	if db := database.GetDB(); db != nil {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			status = http.StatusServiceUnavailable
			deps["db"] = dep{Status: "down", Latency: time.Since(start).String(), Error: "ping failed"}
		} else {
			deps["db"] = dep{Status: "ok", Latency: time.Since(start).String()}
		}
	} else {
		status = http.StatusServiceUnavailable
		deps["db"] = dep{Status: "down", Latency: time.Since(start).String(), Error: "not initialized"}
	}

	// Redis
	start = time.Now()
	if rc := database.GetRedisClient(); rc != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := rc.Ping(ctx).Err(); err != nil {
			status = http.StatusServiceUnavailable
			deps["redis"] = dep{Status: "down", Latency: time.Since(start).String(), Error: err.Error()}
		} else {
			deps["redis"] = dep{Status: "ok", Latency: time.Since(start).String()}
		}
	} else {
		status = http.StatusServiceUnavailable
		deps["redis"] = dep{Status: "down", Latency: time.Since(start).String(), Error: "not initialized"}
	}

	overall := "ok"
	if status != http.StatusOK {
		overall = "degraded"
	}
	c.JSON(status, gin.H{
		"status":       overall,
		"dependencies": deps,
	})
}

// CheckHandler is kept for backward compatibility; it maps to readiness.
func CheckHandler(c *gin.Context) {
	ReadyHandler(c)
}
