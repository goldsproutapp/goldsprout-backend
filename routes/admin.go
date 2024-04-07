package routes

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker/auth"
	"github.com/patrickjonesuk/investment-tracker/constants"
	"github.com/patrickjonesuk/investment-tracker/email"
	"github.com/patrickjonesuk/investment-tracker/middleware"
	"github.com/patrickjonesuk/investment-tracker/models"
	"github.com/patrickjonesuk/investment-tracker/request"
	"gorm.io/gorm"
)

func InviteUser(ctx *gin.Context) {
	user := middleware.GetUser(ctx)
    db := middleware.GetDB(ctx)
	if !user.IsAdmin {
		request.Forbidden(ctx)
		return
	}
    var body models.UserInvitationRequest
    err := ctx.BindJSON(&body)
    if err != nil {
        request.BadRequest(ctx)
        return
    }
    var existingUser models.User
    res := db.Where(&models.User{Email: body.Email}).First(&existingUser)
    if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
        request.BadRequest(ctx) // TODO: give actual information to user
    }
    invitedUser := models.User{
        Email: body.Email,
        FirstName: body.FirstName,
        LastName: body.LastName,
        InvitationToken: auth.GenerateUID(constants.TOKEN_LENGTH),

        IsAdmin: false,
        PasswordHash: "",
        TokenHash: "",
        ClientOpts: "",
    }
    db.Save(&invitedUser)
    email.SendInvitation(body.Email, user, invitedUser.InvitationToken)
    request.Created(ctx, invitedUser)
}

func RegisterAdminRoutes(router *gin.RouterGroup) {
    router.POST("/invite", middleware.Authenticate(), InviteUser)
}
