package router

import (
	controller "easymail/cmd/admin/controller"
	middleware "easymail/cmd/admin/middleware"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"log"
	"path/filepath"
	"strings"
)

func New() *gin.Engine {
	r := gin.New()

	store := cookie.NewStore([]byte("8HVP0sYJN8Izlsyn"))
	r.Use(sessions.Sessions("easymail", store))

	// middlewares
	r.Use(gin.Recovery())
	r.Use(middleware.Access())

	// heartbeat check
	r.GET("/check", func(ctx *gin.Context) {
		count := 0
		sess := sessions.Default(ctx)
		if sess.Get("counter") == nil {
			sess.Set("counter", 1)
			sess.Save()
		} else {
			count = sess.Get("counter").(int)
			sess.Set("counter", count+1)
			sess.Save()
		}
		ctx.JSON(200, gin.H{
			"count": count,
		})
		return
	})

	// load template
	r.HTMLRender = loadTemplates("template")

	// static file serve
	r.Static("/static", "./static")

	// free to go
	homeController := controller.HomeController{}
	r.GET("/favicon.ico", homeController.Favicon)
	r.GET("/login", homeController.Login)
	r.POST("/login", homeController.Login)
	r.GET("/captcha", homeController.Captcha)
	r.POST("/captcha", homeController.Captcha)

	// authorized
	r.Use(middleware.Authorization())
	homeGroup := r.Group("/")
	{
		homeGroup.GET("/", homeController.Dashboard)
		homeGroup.GET("/dashboard", homeController.Dashboard)
		homeGroup.GET("/logout", homeController.Logout)
		homeGroup.GET("/profile", homeController.Profile)
		homeGroup.GET("/password", homeController.ChangePassword)
		homeGroup.POST("/password", homeController.ChangePassword)
	}

	accountController := controller.AccountController{}
	accountGroup := r.Group("/account")
	{
		accountGroup.GET("/domain/index", accountController.IndexDomain)
		accountGroup.POST("/domain/index", accountController.IndexDomain)
		accountGroup.GET("/domain/active", accountController.ToggleDomainActive)
		accountGroup.GET("/domain/delete", accountController.DeleteDomain)
		accountGroup.POST("/domain/create", accountController.CreateDomain)

		accountGroup.GET("/index", accountController.IndexAccount)
		accountGroup.POST("/index", accountController.IndexAccount)
		accountGroup.POST("/create", accountController.CreateAccount)
		accountGroup.GET("/active", accountController.ToggleAccountActive)
		accountGroup.GET("/delete", accountController.DeleteAccount)
		accountGroup.POST("/edit", accountController.EditAccount)
	}

	return r
}

func loadTemplates(templatesDir string) multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	layouts, err := filepath.Glob(templatesDir + "/layout/*")
	if err != nil {
		log.Panic(err)
	}
	singles, err := filepath.Glob(templatesDir + "/single/*")
	if err != nil {
		log.Panic(err)
	}
	includes, err := filepath.Glob(templatesDir + "/**/*")

	if err != nil {
		log.Panic(err)
	}
	for _, include := range includes {
		if strings.Index(include, "/layout/") != -1 {
			continue
		}

		if strings.Index(include, "/single/") != -1 {
			continue
		}

		layoutCopy := make([]string, len(layouts))
		copy(layoutCopy, layouts)
		files := append(layoutCopy, include)
		log.Printf("load template: %s\n", filepath.Base(include))
		log.Printf("load files: %s\n", files)
		r.AddFromFiles(filepath.Base(include), files...)
	}

	for _, s := range singles {
		log.Printf("load single file: %s %s\n", filepath.Base(s), s)
		r.AddFromFiles(filepath.Base(s), []string{s}...)
	}

	return r
}
