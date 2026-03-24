package health

import (
	"context"
	"easymail/internal/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CheckHandler(c *gin.Context) {
	status := http.StatusOK
	payload := gin.H{
		"status": "ok",
		"db":     "ok",
		"redis":  "ok",
	}

	if db := database.GetDB(); db != nil {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			status = http.StatusServiceUnavailable
			payload["status"] = "degraded"
			payload["db"] = "down"
		}
	} else {
		status = http.StatusServiceUnavailable
		payload["status"] = "degraded"
		payload["db"] = "down"
	}

	if rc := database.GetRedisClient(); rc != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if rc.Ping(ctx).Err() != nil {
			status = http.StatusServiceUnavailable
			payload["status"] = "degraded"
			payload["redis"] = "down"
		}
	} else {
		status = http.StatusServiceUnavailable
		payload["status"] = "degraded"
		payload["redis"] = "down"
	}

	c.JSON(status, payload)
}
