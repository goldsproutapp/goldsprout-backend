package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker/auth"
	"github.com/patrickjonesuk/investment-tracker/database"
	"github.com/patrickjonesuk/investment-tracker/middleware"
	"github.com/patrickjonesuk/investment-tracker/models"
	"github.com/patrickjonesuk/investment-tracker/util"
)

func GetUserInfo(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	uids := auth.GetAllowedUsers(user, true, false)
	var users []models.User
	if user.IsAdmin {
		users = database.GetAllUsers(db)
	} else {
		db.Model(&models.User{}).Where("id IN ?", uids).Find(&users)
        users = append(users, user)
	}
	ctx.JSON(http.StatusOK, users)
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

func RegisterUserRoutes(router *gin.RouterGroup) {
	router.GET("/users", middleware.Authenticate("AccessPermissions"), GetUserInfo)
	router.GET("/uservisibility", middleware.Authenticate("AccessPermissions"), GetUserVisibility)
}
