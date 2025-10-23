package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/middleware"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/request/response"
)

func GetAllProviders(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	providers := database.GetProviders(db)
	ctx.JSON(http.StatusOK, providers)
}

func UpdateProvider(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	var body models.ProviderUpdateRequest
	err := ctx.BindJSON(&body)
	if err != nil {
		response.BadRequest(ctx)
		return
	}
	if !user.IsAdmin { // TODO: this seems reasonable for now, but perhaps more granular permissions would be better
		response.Forbidden(ctx)
		return
	}
	db.Save(&(body.Provider))
}

func RegisterProviderRoutes(router *gin.RouterGroup) {
	router.GET("/providers", middleware.Authenticate(), GetAllProviders)
	router.PUT("/providers", middleware.Authenticate(), UpdateProvider)
}
