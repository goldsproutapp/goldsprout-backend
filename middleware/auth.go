package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/auth"
	"github.com/goldsproutapp/goldsprout-backend/config"
	"github.com/goldsproutapp/goldsprout-backend/constants"
	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/request"
	"github.com/goldsproutapp/goldsprout-backend/util"
)

const CtxUserInfoKey = "UserInfo"
const CtxSessionKey = "SessionInfo"

func Authenticate(preload ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		db := GetDB(ctx)

		token := ctx.GetHeader("Authorization")
		var session models.Session
		var user models.User
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "No authentication provided."})
			return
		} else if token == fmt.Sprintf("Bearer %s", constants.DEMO_USER_AUTH_TOKEN) {
			if !config.DemoModeEnabled() || (ctx.Request.Method != "GET" && ctx.Request.Method != "") {
				request.Forbidden(ctx)
				return
			}
			user = database.GetDemoUser(db)
			session = models.Session{
				UserID:        user.ID,
				Client:        util.FormatUA(ctx.Request.UserAgent()),
				IsDemoSession: true,
			}
		} else {
			var hasPrefix bool
			token, hasPrefix = strings.CutPrefix(token, "Bearer ")
			var err error
			session, err = auth.AuthenticateToken(db, token)
			if err != nil || !hasPrefix {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid authentication provided."})
				return
			}
			user, err = auth.UserForSession(db, session, preload...)
			if err != nil || !hasPrefix {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid authentication provided."})
				return
			}
		}

		ctx.Set(CtxUserInfoKey, user)
		ctx.Set(CtxSessionKey, session)
		ctx.Next()
	}
}

func GetUser(ctx *gin.Context) models.User {
	val := ctx.MustGet(CtxUserInfoKey)
	if auth, ok := val.(models.User); ok {
		return auth
	}
	return models.User{}
}

func GetSession(ctx *gin.Context) models.Session {
	val := ctx.MustGet(CtxSessionKey)
	if session, ok := val.(models.Session); ok {
		return session
	}
	return models.Session{}
}
