package routes

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker/auth"
	"github.com/patrickjonesuk/investment-tracker/middleware"
	"github.com/patrickjonesuk/investment-tracker/models"
	"github.com/patrickjonesuk/investment-tracker/request"
	"gorm.io/gorm"
)

func Login(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	header := ctx.GetHeader("Authorization")
	parts := strings.Split(header, ":")
	if len(parts) != 2 {
		request.BadRequest(ctx)
		return
	}
	user, err := auth.AuthenticateUnamePw(db, parts[0], parts[1])
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "invalid username or password"})
		return
	}
	token := auth.CreateToken(db, user)
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"data":    user,
	})
}

func AcceptInvitation(ctx *gin.Context) {
    db := middleware.GetDB(ctx)
    var body models.UserInvitationAccept
    err := ctx.BindJSON(&body)
    if err != nil {
        request.BadRequest(ctx)
        return
    }
    var user models.User
    res := db.Where(models.User{InvitationToken: body.Token, PasswordHash: ""}).First(&user)
    if errors.Is(res.Error, gorm.ErrRecordNotFound) {
        request.BadRequest(ctx)
        return
    }
    user.PasswordHash = auth.HashAndSalt(body.Password)
    token := auth.CreateToken(db, user)
    request.Created(ctx, gin.H{
        "token": token,
        "data": user,
    })
}

func RegisterAuthRoutes(router *gin.RouterGroup) {
	router.POST("/login", Login)
    router.POST("/acceptInvitation", AcceptInvitation)
}

