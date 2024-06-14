package controller

import (
	"easymail/internal/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type ConfigureController struct{}

func (*ConfigureController) Node(c *gin.Context) {
	ids := c.Param("id")
	id, err := strconv.ParseUint(ids, 10, 64)
	if err != nil {
		c.HTML(http.StatusOK, "single_error.html", gin.H{
			"error": err.Error(),
		})
	}

	node, err := model.GetConfigureByID(uint(id))
	if err != nil {
		c.HTML(http.StatusOK, "single_error.html", gin.H{
			"error": err.Error(),
		})
	}
	subNodes, err := model.GetSubConfigureByParentId(uint(id))

	// get root configure nodes
	if c.Request.Method == "GET" {
		sess := sessions.Default(c)
		username := sess.Get("userName")
		menu := createMenu()

		c.HTML(http.StatusOK, "configure_node.html", gin.H{
			"title":    "Configure Management - Easymail",
			"username": username,
			"module":   "configure",
			"menu":     menu,
			"page":     node.Name,
			"id":       node.ID,
			"subNodes": subNodes,
		})
		return
	} else if c.Request.Method == "POST" {
		var req model.ConfigureNodeRequest
		if err = c.BindJSON(&req); err != nil {
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

		total, nodes, err := model.GetConfigureByParentId(uint(id), req.SubNodeID, req.Keyword, orderField, orderDir, req.Start, req.Length)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search error"})
			return
		}
		responseData := make([]model.ConfigureNodeResponse, 0)
		for _, detailNode := range nodes {
			responseData = append(responseData, model.ConfigureNodeResponse{
				ID:          detailNode.ID,
				TopName:     node.Name,
				SubName:     detailNode.Parent.Name,
				Name:        detailNode.Name,
				Value:       detailNode.Value,
				DataType:    uint(detailNode.DataType),
				CreateTime:  detailNode.CreateTime.Format("2006-01-02 15:04:05"),
				Private:     detailNode.Private,
				Description: detailNode.Description,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"draw":            req.Draw,
			"recordsTotal":    total,
			"recordsFiltered": total,
			"data":            responseData,
		})
	}
}

type configureEditRequest struct {
	Value       string `json:"value"`
	Description string `json:"description"`
}

func (*ConfigureController) Edit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	node, err := model.GetConfigureByID(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	var req configureEditRequest
	if err = c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	node.UpdateTime = time.Now()
	node.Value = req.Value
	node.Description = req.Description
	if err := node.Save(); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "success"})
}
