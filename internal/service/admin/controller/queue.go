package controller

import (
	"easymail/internal/postfix/queue"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type QueueController struct{}

func (qc *QueueController) Index(c *gin.Context) {
	if c.Request.Method == "GET" {
		sess := sessions.Default(c)
		username := sess.Get("userName")
		menu := createMenu()
		c.HTML(http.StatusOK, "postfix_queue.html", gin.H{
			"title":    "Mail Queue Of Postfix - Easymail",
			"username": username,
			"module":   "postfix",
			"page":     "queue",
			"menu":     menu,
		})
		return
	} else if c.Request.Method == "POST" {
		queues, err := queue.Dump()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": queues,
		})
	}
}

func (qc *QueueController) View(c *gin.Context) {
	id := c.Query("id")
	result, err := queue.View(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

type BatchQueueRequest struct {
	IDS []string `json:"ids"`
}

func (qc *QueueController) Flush(c *gin.Context) {
	if c.Request.Method == "GET" {
		id := c.Query("id")
		result, err := queue.Flush(id)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	} else if c.Request.Method == "POST" {
		var req BatchQueueRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}
		data := make([]string, 0)

		for _, id := range req.IDS {
			result, err := queue.Flush(id)
			if err != nil {
				data = append(data, err.Error())
			}
			data = append(data, result)
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": strings.Join(data, "\n")})
	}
}

func (qc *QueueController) Delete(c *gin.Context) {
	if c.Request.Method == "GET" {
		id := c.Query("id")
		result, err := queue.Delete(id)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	} else if c.Request.Method == "POST" {
		var req BatchQueueRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}
		data := make([]string, 0)

		for _, id := range req.IDS {
			result, err := queue.Delete(id)
			if err != nil {
				data = append(data, err.Error())
			}
			data = append(data, result)
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": strings.Join(data, "\n")})
	}
}
