package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker-backend/database"
	"github.com/patrickjonesuk/investment-tracker-backend/middleware"
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

func RegisterMiscRoutes(router *gin.RouterGroup) {
    router.GET("/regions", middleware.Authenticate(), GetAllRegions)
    router.GET("/sectors", middleware.Authenticate(), GetAllSectors)
}
