package router

import (
	"easymail/internal/easylog"
	"easymail/internal/service/admin/controller"
	"easymail/internal/service/admin/middleware"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"log"
	"path"
	"path/filepath"
	"strings"
)

func New(_log *easylog.Logger, root, cookiePassword, cookieTag string) *gin.Engine {
	r := gin.New()

	store := cookie.NewStore([]byte(cookiePassword))
	r.Use(sessions.Sessions(cookieTag, store))

	// middlewares
	r.Use(gin.Recovery())
	r.Use(middleware.Access(_log))

	// heartbeat check
	r.GET("/check", func(c *gin.Context) {
		count := 0
		sess := sessions.Default(c)
		if sess.Get("counter") == nil {
			sess.Set("counter", 1)
			sess.Save()
		} else {
			count = sess.Get("counter").(int)
			sess.Set("counter", count+1)
			sess.Save()
		}
		c.JSON(200, gin.H{
			"count": count,
		})
		return
	})

	// load template
	r.HTMLRender = loadTemplates(path.Join(root, "template"))

	// static file serve
	r.Static("/static", path.Join(root, "static"))

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
		accountGroup.GET("/active", accountController.ToggleAccount)
		accountGroup.GET("/delete", accountController.DeleteAccount)
		accountGroup.POST("/edit", accountController.EditAccount)
	}

	mailLogController := controller.MailLogController{}
	queueController := controller.QueueController{}
	postfixGroup := r.Group("/postfix")
	{
		postfixGroup.GET("/maillog/index", mailLogController.Index)
		postfixGroup.POST("/maillog/index", mailLogController.Index)
		postfixGroup.GET("/queue/index", queueController.Index)
		postfixGroup.POST("/queue/index", queueController.Index)
		postfixGroup.GET("/queue/view", queueController.View)
		postfixGroup.GET("/queue/flush", queueController.Flush)
		postfixGroup.GET("/queue/delete", queueController.Delete)
		postfixGroup.POST("/queue/flush", queueController.Flush)
		postfixGroup.POST("/queue/delete", queueController.Delete)
	}

	configureController := controller.ConfigureController{}
	configureGroup := r.Group("/configure")
	{
		configureGroup.GET("/node/:id", configureController.Node)
		configureGroup.POST("/node/:id", configureController.Node)
		configureGroup.POST("/node/edit/:id", configureController.Edit)
	}

	filterController := controller.FilterController{}
	filterGroup := r.Group("/filter")
	{
		filterGroup.GET("/feature", filterController.Feature)
		filterGroup.POST("/feature", filterController.Feature)
		filterGroup.POST("/feature/field", filterController.IndexField)
		filterGroup.POST("/feature/metric", filterController.IndexMetric)
		filterGroup.POST("/feature/metric/create", filterController.CreateMetric)
		filterGroup.GET("/feature/metric/toggle", filterController.ToggleMetric)
		filterGroup.GET("/feature/metric/edit", filterController.EditMetric)
		filterGroup.POST("/feature/metric/edit", filterController.EditMetric)
		filterGroup.POST("/feature/metric/delete", filterController.DeleteMetric)
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
		r.AddFromFiles(filepath.Base(include), files...)
	}

	for _, s := range singles {
		r.AddFromFiles(filepath.Base(s), []string{s}...)
	}

	return r
}
