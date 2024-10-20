package routes

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/auth"
	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/email"
	"github.com/goldsproutapp/goldsprout-backend/middleware"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/request"
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
		Trusted:      false,
		Active:       false,
		PasswordHash: "",
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
	if !database.Exists(db.Model(&models.User{}).Where("id = ?", body.User).Preload("AccessPermissions").First(&user)) {
		request.NotFound(ctx)
		return
	}
	if user.Trusted != *body.Trusted {
		user.Trusted = *body.Trusted
		db.Save(&user)
	}
	permissionMap := map[uint]models.AccessPermission{}
	for _, perm := range user.AccessPermissions {
		permissionMap[perm.AccessForID] = perm
	}
	updatePermissions := make([]models.AccessPermission, 0)
	toDelete := make([]models.AccessPermission, 0)
	for _, perm := range body.Permissions {
		existing, exists := permissionMap[perm.ForUser]
		if exists && existing.Read == perm.Read && existing.Write == perm.Write && existing.Limited == perm.Limited {
			continue
		}
		if !perm.Read && !perm.Write && !perm.Limited {
			if exists {
				toDelete = append(toDelete, models.AccessPermission{ID: existing.ID})
			} else {
				continue
			}
		} else {
			updatePermissions = append(updatePermissions, models.AccessPermission{
				UserID:      user.ID,
				AccessForID: perm.ForUser,
				Read:        perm.Read,
				Write:       perm.Write,
				Limited:     perm.Limited,
			})
		}
	}
	if len(toDelete) > 0 {
		db.Delete(&toDelete)
	}
	if len(updatePermissions) > 0 {
		db.Save(&updatePermissions)
	}
	request.OK(ctx, user)
}

func MassDelete(ctx *gin.Context) {
	user := middleware.GetUser(ctx)
	if !user.IsAdmin {
		request.Forbidden(ctx)
		return
	}
	var body models.MassDeleteRequest
	if ctx.BindJSON(&body) != nil {
		request.BadRequest(ctx)
		return
	}
	db := middleware.GetDB(ctx)
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.StockSnapshot{})
	if body.Stocks {
		db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.UserStock{})
		db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Stock{})
	}
}

func RegisterAdminRoutes(router *gin.RouterGroup) {
	router.POST("/invite", middleware.Authenticate(), InviteUser)
	router.PUT("/permissions", middleware.Authenticate(), SetPermissions)
	router.POST("/massdelete", middleware.Authenticate(), MassDelete)
}
