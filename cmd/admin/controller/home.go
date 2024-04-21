package controller

import (
	"easymail/internal/account"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

type HomeController struct{}

type loginRequest struct {
	Username string `form:"username" binding:"required,email,min=6"`
	Password string `form:"password" binding:"required,min=6"`
}

func (home HomeController) Favicon(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
	return
}

func (home HomeController) Login(ctx *gin.Context) {
	if ctx.Request.Method == http.MethodGet {
		ctx.HTML(http.StatusOK, "single_login.html", gin.H{})
		return
	} else if ctx.Request.Method == http.MethodPost {
		req := loginRequest{}
		err := ctx.ShouldBind(&req)
		if err != nil {
			ctx.HTML(http.StatusOK, "single_login.html", gin.H{
				"error": "error:" + err.Error(),
			})
			return
		}

		username := req.Username
		password := req.Password

		if len(username) < 3 || len(password) < 6 {
			ctx.HTML(http.StatusOK, "single_login.html", gin.H{
				"username": username,
				"password": password,
				"error":    "username or password are invalid",
			})
			return
		}

		acc, err := account.Authorize(username, password)
		if err != nil {
			ctx.HTML(http.StatusOK, "single_login.html", gin.H{
				"username": username,
				"password": password,
				"error":    "username or password is invalid",
			})
			return
		}

		// set session
		sess := sessions.Default(ctx)
		sess.Set("userID", strconv.Itoa(int(acc.ID)))
		sess.Set("userName", acc.Username)
		sess.Set("domainID", strconv.Itoa(int(acc.DomainID)))
		err = sess.Save()
		if err != nil {
			log.Println("failed to save session:", err)
		}

		ctx.Redirect(http.StatusFound, "/dashboard")
		return
	}
}

func (home HomeController) Logout(ctx *gin.Context) {
	// clean session data
	sess := sessions.Default(ctx)
	sess.Clear()
	err := sess.Save()
	if err != nil {
		log.Println("failed to save session:", err)
	}

	ctx.Redirect(http.StatusFound, "/login")
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

func (home HomeController) Dashboard(ctx *gin.Context) {
	username := "Easymail User"
	sess := sessions.Default(ctx)
	userIDS := sess.Get("userID").(string)
	userID, err := strconv.Atoi(userIDS)
	if err == nil {
		acc, err := account.GetAccountByID(userID)
		if err == nil {
			username = acc.Username
		}
	}
	ctx.HTML(http.StatusOK, "home_dashboard.html", gin.H{
		"title":    "Dashboard of admin - Easymail",
		"username": username,
	})
}
