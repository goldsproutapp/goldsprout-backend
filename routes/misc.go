package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/middleware"
	"github.com/goldsproutapp/goldsprout-backend/request"
)

func GetAllRegions(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	regions := database.GetRegions(db)
	ctx.JSON(http.StatusOK, gin.H{"regions": regions})
}

func GetAllSectors(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	sectors := database.GetSectors(db)
	ctx.JSON(http.StatusOK, gin.H{"sectors": sectors})
}

func GetAllClasses(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	classes := database.GetClasses(db)
	request.OK(ctx, gin.H{"classes": classes})
}

func RegisterMiscRoutes(router *gin.RouterGroup) {
	router.GET("/regions", middleware.Authenticate(), GetAllRegions)
	router.GET("/sectors", middleware.Authenticate(), GetAllSectors)
	router.GET("/classes", middleware.Authenticate(), GetAllClasses)
}
