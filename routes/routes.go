package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/middleware"
	"gorm.io/gorm"
)

func RegisterAllRoutes(router *gin.RouterGroup, db *gorm.DB) {
	router.Use(middleware.Database(db))
	RegisterAuthRoutes(router)

	RegisterStockRoutes(router)
	RegisterSnapshotRoutes(router)
	RegisterProviderRoutes(router)
	RegisterPerformanceRoutes(router)
	RegisterOverviewRoutes(router)
	RegisterSplitRoutes(router)
	RegisterAccountRoutes(router)

	RegisterUserRoutes(router)
	RegisterMiscRoutes(router)
	RegisterAdminRoutes(router)
	RegisterExportRoutes(router)
}
