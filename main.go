package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker-backend/config"
	"github.com/patrickjonesuk/investment-tracker-backend/database"
	"github.com/patrickjonesuk/investment-tracker-backend/middleware"
	"github.com/patrickjonesuk/investment-tracker-backend/routes"
)

func UserInfo(ctx *gin.Context) {
	user := middleware.GetUser(ctx)
	ctx.JSON(http.StatusOK, user)
}

func main() {

	db := database.InitDB()
	database.CreateInitialAdminAccount(db)
	if config.DemoModeEnabled() {
		database.CreateDemoAccount(db)
	}

	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	router.Use(middleware.Database(db))
	routes.RegisterAllRoutes(&router.RouterGroup, db)

	router.Run(config.EnvOrDefault(config.LISTEN_INTERFACE, "0.0.0.0") + ":" + config.EnvOrDefault(config.LISTEN_PORT, "3000"))

}
