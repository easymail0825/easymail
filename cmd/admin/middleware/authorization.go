package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Authorization() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// get session object from ctx

		// check userID and domainID in session
		sess := sessions.Default(ctx)
		if sess == nil {
			ctx.Redirect(http.StatusTemporaryRedirect, "/login")
			return
		}
		userID := sess.Get("userID")
		domainID := sess.Get("domainID")
		if userID == nil || domainID == nil {
			ctx.Redirect(http.StatusTemporaryRedirect, "/login")
			return
		}
		ctx.Next()
	}
}
