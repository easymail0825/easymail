package router

import (
	"easymail/internal/easylog"
	"easymail/internal/service/webmail/controller"
	"easymail/internal/service/webmail/middleware"
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
		homeGroup.GET("/", homeController.Index)
		homeGroup.GET("/logout", homeController.Logout)
		homeGroup.GET("/profile", homeController.Profile)
		homeGroup.GET("/password", homeController.ChangePassword)
		homeGroup.POST("/password", homeController.ChangePassword)
	}

	mailboxController := controller.MailboxController{}
	mailboxGroup := r.Group("/mailbox")
	{
		mailboxGroup.GET("/", mailboxController.Index)
		mailboxGroup.GET("/:folder", mailboxController.Index)
		mailboxGroup.POST("/markread", mailboxController.MarkRead)
		mailboxGroup.GET("/read/:mid", mailboxController.Read)
		mailboxGroup.POST("/attach/:mid", mailboxController.DownloadAttach)
		mailboxGroup.POST("/delete", mailboxController.DeleteMails)
		mailboxGroup.POST("/folder/:folderID", mailboxController.Folder)
		mailboxGroup.GET("/write", mailboxController.Write)
		mailboxGroup.POST("/write", mailboxController.Write)
		mailboxGroup.POST("/done", mailboxController.Done)

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
