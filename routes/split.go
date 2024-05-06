package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker-backend/calculations/performance"
	"github.com/patrickjonesuk/investment-tracker-backend/calculations/split"
	"github.com/patrickjonesuk/investment-tracker-backend/database"
	"github.com/patrickjonesuk/investment-tracker-backend/middleware"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/request"
	"github.com/shopspring/decimal"
)

func Split(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)

	var query models.StockFilterQuery
	if ctx.BindQuery(&query) != nil {
		request.BadRequest(ctx)
		return
	}
	filter := performance.BuildStockFilter(query)

	snapshots := database.FetchPerformanceData(db, user, filter)
	categories := []string{"region", "sector", "provider", "stock"}
	out := map[string]map[string]decimal.Decimal{}
	for _, categoryKey := range categories {
		groups := split.CategoriseSnapshots(snapshots, categoryKey)
		result := split.CalculateSplit(groups)
		out[categoryKey] = result
	}
	request.OK(ctx, out)
}

func RegisterSplitRoutes(router *gin.RouterGroup) {
	router.GET("/split", middleware.Authenticate("AccessPermissions"), Split)
}
