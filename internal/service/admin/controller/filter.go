package controller

import (
	"easymail/internal/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type FilterController struct{}

func (*FilterController) Feature(c *gin.Context) {
	// get root configure nodes
	if c.Request.Method == "GET" {
		sess := sessions.Default(c)
		username := sess.Get("userName")
		menu := createMenu()

		// get filter field list total
		fields, err := model.GetAllFilterField()
		if err != nil {
			c.HTML(http.StatusOK, "single_error.html", gin.H{"error": err.Error()})
			return
		}

		c.HTML(http.StatusOK, "filter_feature.html", gin.H{
			"title":    "filter feature - Easymail",
			"username": username,
			"module":   "filter",
			"menu":     menu,
			"page":     "index",
			"fields":   fields,
		})
		return
	}
}

func (*FilterController) IndexField(c *gin.Context) {
	if c.Request.Method == "POST" {
		var req model.IndexFilterFieldRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": err.Error(),
			})
		}

		orderField := ""
		orderDir := ""
		for _, o := range req.Orders {
			orderField = req.Columns[o.Column].Data
			orderDir = o.Dir
			break
		}

		// get filter field list in page
		total, data, err := model.GetFilterField(orderField, orderDir, req.Start, req.Length)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"draw":            req.Draw,
			"recordsTotal":    total,
			"recordsFiltered": total,
			"data":            data,
		})
	}
}

func (*FilterController) IndexMetric(c *gin.Context) {
	if c.Request.Method == "POST" {
		var req model.IndexFilterMetricRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": err.Error(),
			})
		}

		for _, o := range req.Orders {
			req.OrderField = req.Columns[o.Column].Data
			req.OrderDir = o.Dir
			break
		}

		total, data, err := model.GetFilterMetric(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"draw":            req.Draw,
			"recordsTotal":    total,
			"recordsFiltered": total,
			"data":            data,
		})
	}
}

func (*FilterController) CreateMetric(c *gin.Context) {
	if c.Request.Method == "POST" {
		var req model.CreateFilterMetricRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": err.Error(),
			})
			return
		}
		sess := sessions.Default(c)
		acc := sess.Get("account").(model.Account)
		req.AccountID = acc.ID

		// edit
		if req.ID > 0 {
			err := model.EditFilterMetric(req)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"error": "edit metric failed:" + err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": "edit metric success"})
			return
		}

		err := model.CreateFilterMetric(req)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "create model failed:" + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": "create model success"})
	}
}

func (*FilterController) ToggleMetric(c *gin.Context) {
	ids := c.Query("id")
	if ids == "" {
		c.JSON(http.StatusOK, gin.H{"error": "invalid model id"})
		return
	}

	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "invalid model id"})
		return
	}

	err = model.ToggleFilterMetric(id)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "toggle model active error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "toggle model active success"})
}

func (*FilterController) DeleteMetric(c *gin.Context) {
	ids := c.Query("id")
	if ids == "" {
		c.JSON(http.StatusOK, gin.H{"error": "invalid model id"})
		return
	}
	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "invalid model id"})
		return
	}

	err = model.DeleteFilterMetric(id)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "delete metric error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "delete metric success"})
}

func (*FilterController) EditMetric(c *gin.Context) {
	if c.Request.Method == "GET" {
		ids := c.Query("id")
		if ids == "" {
			c.JSON(http.StatusOK, gin.H{"error": "invalid metric id"})
			return
		}
		id, err := strconv.ParseInt(ids, 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "invalid metric id"})
			return
		}
		metric, err := model.GetFilterMetricByID(id)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "get metric error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": metric, "success": "ok"})
	}
}
