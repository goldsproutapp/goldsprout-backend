package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/auth"
	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/middleware"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/request"
	"github.com/goldsproutapp/goldsprout-backend/util"
)

func GetUserInfo(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	uids := auth.GetAllowedUsers(user, true, false, true)
	if user.IsAdmin {
		request.OK(ctx, database.GetAllUsers(db, "AccessPermissions"))
	} else {
		var users []models.User
		db.Model(&models.User{}).Where("id IN ?", uids).Find(&users)
		request.OK(ctx, util.Map(users, models.User.PublicInfo))
	}
}

func GetUserVisibility(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	var aps []models.AccessPermission
	db.Model(&models.AccessPermission{}).Where("access_for_id = ?", user.ID).Find(&aps)
	var users []models.User
	db.Model(&models.User{}).Where("id IN ? OR is_admin = true", util.Map(aps, func(ap models.AccessPermission) uint {
		return ap.UserID
	})).Find(&users)
	ctx.JSON(http.StatusOK, util.Map(users, models.User.PublicInfo))
}

func UpdateUserInfo(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	var body models.UserUpdateInfo
	if ctx.BindJSON(&body) != nil {
		request.BadRequest(ctx)
		return
	}
	user.ApplyUpdate(body)
	db.Save(&user)
	request.OK(ctx, user)
}

func RegisterUserRoutes(router *gin.RouterGroup) {
	router.GET("/users", middleware.Authenticate("AccessPermissions"), GetUserInfo)
	router.GET("/uservisibility", middleware.Authenticate("AccessPermissions"), GetUserVisibility)
	router.PATCH("/user", middleware.Authenticate(), UpdateUserInfo)
}
