package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/middleware"
)

func Overview(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	overview := database.GetOverview(db, user)
	ctx.JSON(http.StatusOK, overview)

}

func RegisterOverviewRoutes(router *gin.RouterGroup) {
	router.GET("/overview", middleware.Authenticate("AccessPermissions"), Overview)
}
