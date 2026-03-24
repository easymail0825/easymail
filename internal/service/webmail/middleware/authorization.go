package middleware

import (
	sessionkey "easymail/internal/application/session"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Authorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := sessions.Default(c)
		if v := sess.Get(sessionkey.KeyUserID); v == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
		} else {
			c.Next()
		}
	}
}
