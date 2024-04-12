package routes

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker-backend/auth"
	"github.com/patrickjonesuk/investment-tracker-backend/email"
	"github.com/patrickjonesuk/investment-tracker-backend/middleware"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/request"
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
		request.Conflict(ctx)
	}
	invitedUser := models.User{
		Email:           body.Email,
		FirstName:       body.FirstName,
		LastName:        body.LastName,
		InvitationToken: auth.GenerateToken(),

		IsAdmin:      false,
		Active:       false,
		PasswordHash: "",
		TokenHash:    "",
		ClientOpts:   "",
	}
	db.Save(&invitedUser)
	email.SendInvitation(body.Email, user, invitedUser.InvitationToken)
	request.Created(ctx, invitedUser)
}

func SetPermissions(ctx *gin.Context) {
	admin := middleware.GetUser(ctx)
	db := middleware.GetDB(ctx)
	if !admin.IsAdmin {
		request.Forbidden(ctx)
		return
	}
	var body models.SetPermissionsRequest
	if ctx.BindJSON(&body) != nil {
		request.BadRequest(ctx)
		return
	}
	var user models.User
	res := db.Where("user_id = ?", body.User).Preload("AccessPermissions").First(&user)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		request.NotFound(ctx)
		return
	}
	permissionMap := map[uint]models.AccessPermission{}
	for _, perm := range user.AccessPermissions {
		permissionMap[perm.AccessForID] = perm
	}
	updatePermissions := make([]models.AccessPermission, 0)
	toDelete := make([]models.AccessPermission, 0)
	for _, perm := range body.Permissions {
		existing, exists := permissionMap[perm.ForUser]
		if exists && existing.Read == perm.Read && existing.Write == perm.Write {
			continue
		}
		if !perm.Read && !perm.Write {
			if exists {
				toDelete = append(toDelete, models.AccessPermission{UserID: user.ID, AccessForID: perm.ForUser})
			} else {
				continue
			}
		} else {
			updatePermissions = append(updatePermissions, models.AccessPermission{
				UserID:      user.ID,
				AccessForID: perm.ForUser,
				Read:        perm.Read,
				Write:       perm.Write,
			})
		}
	}
	db.Delete(&toDelete)
	db.Save(&updatePermissions)
	request.OK(ctx, gin.H{})
}

func RegisterAdminRoutes(router *gin.RouterGroup) {
	router.POST("/invite", middleware.Authenticate(), InviteUser)
    router.PUT("/permissions", middleware.Authenticate(), SetPermissions)
}
