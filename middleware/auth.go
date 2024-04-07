package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker/auth"
	"github.com/patrickjonesuk/investment-tracker/models"
)

const CtxAuthKey = "AuthInfo"

func Authenticate(preload ...string) gin.HandlerFunc {
    return func (ctx *gin.Context) {
        db := GetDB(ctx)

        token := ctx.GetHeader("Authorization")
        if token == "" {
            ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "No authentication provided."})
            return
        }
        var hasPrefix bool
        token, hasPrefix = strings.CutPrefix(token, "Bearer ")
        user, err := auth.AuthenticateToken(db, token, preload...)
        if err != nil || !hasPrefix {
            ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid authentication provided."})
            return
        }

        ctx.Set(CtxAuthKey, user)
        ctx.Next()
    }
}

func GetUser(ctx *gin.Context) models.User {
    val := ctx.MustGet(CtxAuthKey)
    if auth, ok := val.(models.User); ok {
        return auth
    }
    return models.User{}
}
