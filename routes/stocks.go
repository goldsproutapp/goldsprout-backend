package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker-backend/auth"
	"github.com/patrickjonesuk/investment-tracker-backend/database"
	"github.com/patrickjonesuk/investment-tracker-backend/middleware"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/request"
)

func GetAllStocks(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	userStocks := database.GetVisibleStockList(user, db)
	ctx.JSON(http.StatusOK, userStocks)
}

func UpdateStock(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	var body models.StockUpdateRequest
	err := ctx.BindJSON(&body)
	if err != nil {
		request.BadRequest(ctx)
		return
	}
	if !user.IsAdmin {
		/*
					If the user is not an admin, then they can only modify the stock
					If they have write permissions for every user who holds it
			        (or they are the only user who holds it)
		*/
		uids, err := database.GetUsersHoldingStock(db, body.Stock.ID)
		if err != nil {
			request.Forbidden(ctx)
			return
		}
		for _, uid := range uids {
			if uid != user.ID && !auth.HasAccessPerm(user, uid, false, true) {
				request.Forbidden(ctx)
				return
			}
		}
	}
	db.Save(&(body.Stock))

}

func RegisterStockRoutes(router *gin.RouterGroup) {
	router.GET("/stocks", middleware.Authenticate("AccessPermissions"), GetAllStocks)
	router.PUT("/stocks", middleware.Authenticate("AccessPermissions"), UpdateStock)
}
