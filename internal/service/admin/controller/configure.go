package controller

import (
	"easymail/internal/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ConfigureController struct{}

type ConfigureData struct {
	ID         uint   `json:"id"`
	SubName    string `json:"subName"`
	Name       string `json:"name"`
	Value      string `json:"value"`
	DataType   uint   `json:"dataType"`
	Private    bool   `json:"private"`
	CreateTime string `json:"createTime"`
	Describe   string `json:"describe"`
}

type ConfigureResponse struct {
	Data []ConfigureData `json:"data"`
}

type ConfigureRequest struct {
	Value    string `json:"value"`
	Describe string `json:"describe"`
}

func (*ConfigureController) Node(c *gin.Context) {
	ids := c.Param("id")
	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		c.HTML(http.StatusOK, "single_error.html", gin.H{
			"error": err.Error(),
		})
	}

	// get root configure nodes
	if c.Request.Method == "GET" {
		sess := sessions.Default(c)
		username := sess.Get("userName")

		menu := createMenu()

		subNodes, err := model.GetConfigureByParentId(id)
		responseData := ConfigureResponse{
			Data: make([]ConfigureData, 0),
		}
		node, err := model.GetConfigureByID(id)
		if err != nil {
			c.HTML(http.StatusOK, "single_error.html", gin.H{
				"error": err.Error(),
			})
		}
		for _, node := range subNodes {
			//responseData.Data = append(responseData.Data, ConfigureData{
			//	ID:         node.ID,
			//	Name:       node.Name,
			//	Value:      node.Value,
			//	DataType:   uint(node.DataType),
			//	CreateTime: node.CreateTime,
			//	Private:    node.Private,
			//})
			detailNodes, err := model.GetConfigureByParentId(int64(node.ID))
			if err == nil {
				for _, detailNode := range detailNodes {
					responseData.Data = append(responseData.Data, ConfigureData{
						ID:         detailNode.ID,
						SubName:    node.Name,
						Name:       detailNode.Name,
						Value:      detailNode.Value,
						DataType:   uint(detailNode.DataType),
						CreateTime: detailNode.CreateTime.Format("2006-01-02 15:04:05"),
						Private:    detailNode.Private,
						Describe:   detailNode.Describe,
					})
				}
			}
		}
		c.HTML(http.StatusOK, "configure_node.html", gin.H{
			"title":    "Configure Management - Easymail",
			"username": username,
			"module":   "configure",
			"menu":     menu,
			"data":     responseData.Data,
			"page":     node.Name,
		})
		return
	} else if c.Request.Method == "POST" {
		node, err := model.GetConfigureByID(id)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": err.Error(),
			})
		}

		var req ConfigureRequest
		if err = c.BindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": err.Error(),
			})
		}

		node.Value = req.Value
		node.Describe = req.Describe
		if err = model.UpdateConfigure(*node); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": err.Error(),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"success": "Update success",
		})

	}
}
