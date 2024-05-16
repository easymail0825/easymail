package controller

import (
	"easymail/internal/maillog"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"html"
	"net/http"
	"time"
)

type MailLogController struct{}

type MailLogIndexRequest struct {
	DataTableRequest
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Keyword     string `json:"keyword"`
	SearchField int    `json:"searchField"`
}
type MailLogIndexResponse struct {
	ID        int64     `json:"id"`
	LogTime   time.Time `json:"logTime"`
	SessionID string    `json:"sessionID"`
	Process   string    `json:"process"`
	Message   string    `json:"message"`
}

func (a *MailLogController) Index(c *gin.Context) {
	if c.Request.Method == "GET" {
		sess := sessions.Default(c)
		username := sess.Get("userName")
		menu := createMenu()
		c.HTML(http.StatusOK, "postfix_maillog.html", gin.H{
			"title":    "Mail Logs Of Postfix - Easymail",
			"username": username,
			"module":   "postfix",
			"page":     "mailLog",
			"menu":     menu,
		})
		return
	} else if c.Request.Method == "POST" {
		var req MailLogIndexRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		orderField := ""
		orderDir := ""
		for _, o := range req.Orders {
			orderField = req.Columns[o.Column].Data
			orderDir = o.Dir
			break
		}

		var startTime, endTime time.Time
		var err error
		if req.StartDate != "" && req.EndDate != "" {
			startTime, err = time.ParseInLocation("2006-01-02", req.StartDate, time.Local)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
				return
			}
			endTime, err = time.ParseInLocation("2006-01-02", req.EndDate, time.Local)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
				return
			}
		}

		total, logs, err := maillog.Index(startTime, endTime, req.SearchField, req.Keyword, orderField, orderDir, req.Start, req.Length)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search mail logs error"})
			return
		}

		data := make([]MailLogIndexResponse, 0)
		for _, l := range logs {
			data = append(data, MailLogIndexResponse{
				ID:        l.ID,
				LogTime:   l.LogTime,
				SessionID: l.SessionID,
				Process:   l.Process,
				Message:   html.EscapeString(l.Message),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"draw":            req.Draw,
			"recordsTotal":    total,
			"recordsFiltered": total,
			"data":            data,
		})
	}
}
