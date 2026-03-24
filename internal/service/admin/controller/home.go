package controller

import (
	"context"
	sessionkey "easymail/internal/application/session"
	"easymail/internal/identity"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type HomeController struct{}

var adminIdentityService = identity.NewService()

type loginRequest struct {
	Username string `form:"username" binding:"required,min=6"`
	Password string `form:"password" binding:"required,min=6"`
}

func (home HomeController) Favicon(c *gin.Context) {
	c.Status(http.StatusOK)
	return
}

func (home HomeController) Login(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, "single_login.html", gin.H{})
		return
	} else if c.Request.Method == http.MethodPost {
		req := loginRequest{}
		err := c.ShouldBind(&req)
		if err != nil {
			c.HTML(http.StatusOK, "single_login.html", gin.H{
				"error": "error:" + err.Error(),
			})
			return
		}

		username := req.Username
		password := req.Password

		if len(username) < 3 || len(password) < 6 {
			c.HTML(http.StatusOK, "single_login.html", gin.H{
				"username": username,
				"password": password,
				"error":    "username or password are invalid",
			})
			return
		}

		acc, err := adminIdentityService.Authenticate(context.Background(), username, password)
		if err != nil {
			c.HTML(http.StatusOK, "single_login.html", gin.H{
				"username": username,
				"password": password,
				"error":    "username or password is invalid",
			})
			return
		}

		// set session
		sess := sessions.Default(c)
		sess.Set(sessionkey.KeyAdminAccount, acc.Username)
		sess.Set("userName", acc.Username)
		err = sess.Save()
		if err != nil {
			log.Println("failed to save session:", err)
		}

		c.Redirect(http.StatusFound, "/dashboard")
		return
	}
}

func (home HomeController) Logout(c *gin.Context) {
	// clean session data
	sess := sessions.Default(c)
	sess.Clear()
	err := sess.Save()
	if err != nil {
		log.Println("failed to save session:", err)
	}

	c.Redirect(http.StatusFound, "/login")
	return
}

func (home HomeController) Captcha(c *gin.Context) {
	return
}

func (home HomeController) Profile(context *gin.Context) {
	return
}

func (home HomeController) ChangePassword(context *gin.Context) {
	return
}

func (home HomeController) Dashboard(c *gin.Context) {
	sess := sessions.Default(c)
	username, _ := sess.Get(sessionkey.KeyAdminAccount).(string)
	menu := createMenu()
	c.HTML(http.StatusOK, "home_dashboard.html", gin.H{
		"title":    "Dashboard of admin - Easymail",
		"username": username,
		"module":   "dashboard",
		"page":     "home",
		"menu":     menu,
	})
}
