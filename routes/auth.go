package routes

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker-backend/auth"
	"github.com/patrickjonesuk/investment-tracker-backend/config"
	"github.com/patrickjonesuk/investment-tracker-backend/constants"
	"github.com/patrickjonesuk/investment-tracker-backend/database"
	"github.com/patrickjonesuk/investment-tracker-backend/middleware"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/request"
	"gorm.io/gorm"
)

func Login(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")
	parts := strings.Split(header, ":")
	if len(parts) != 2 {
		request.BadRequest(ctx)
		return
	}
	db := middleware.GetDB(ctx)
	var user models.User
	var token string
	if parts[0] == "demo" && parts[1] == "demo" && config.DemoModeEnabled() {
		user = database.GetDemoUser(db)
		token = constants.DEMO_USER_AUTH_TOKEN
	} else {
		var err error
		user, err = auth.AuthenticateUnamePw(db, parts[0], parts[1])
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "invalid username or password"})
			return
		}
		token = auth.CreateToken(db, user, ctx.Request.UserAgent())
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"data":    user,
	})
}

func Logout(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	session := middleware.GetSession(ctx)
	if !session.IsDemoSession {
		db.Delete(&session)
	}
	request.NoContent(ctx)
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
	res := db.Where(models.User{InvitationToken: body.Token, PasswordHash: "", Active: false}).First(&user)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		request.BadRequest(ctx)
		return
	}
	user.PasswordHash = auth.HashAndSalt(body.Password)
	user.Active = true
	token := auth.CreateToken(db, user, ctx.Request.UserAgent())
	request.Created(ctx, gin.H{
		"token": token,
		"data":  user,
	})
}

func ChangePassword(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	var body models.PasswordChangeRequest
	if ctx.BindJSON(&body) != nil {
		request.BadRequest(ctx)
		return
	}
	if !auth.ValidatePassword(body.OldPassword, user.PasswordHash) {
		request.Forbidden(ctx)
		return
	}
	user.PasswordHash = auth.HashAndSalt(body.NewPassword)
	db.Save(&user)
	ctx.Status(http.StatusOK)
}

func RegisterAuthRoutes(router *gin.RouterGroup) {
	router.POST("/login", Login)
	router.POST("/logout", middleware.Authenticate(), Logout)
	router.POST("/acceptInvitation", AcceptInvitation)
	router.PATCH("/changepassword", middleware.Authenticate(), ChangePassword)
}
