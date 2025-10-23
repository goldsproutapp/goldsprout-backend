package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/calculations/trends"
	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/middleware"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/request"
	"github.com/goldsproutapp/goldsprout-backend/request/response"
)

func Trends(ctx *gin.Context) {
	var query models.PerformanceRequestQuery
	err := ctx.BindQuery(&query)
	if err != nil {
		response.BadRequest(ctx)
		return
	}
	info := trends.PerformanceQueryInfo{
		TargetKey:  query.Of,
		AgainstKey: query.For,
		TimeKey:    query.Over,
		MetricKey:  query.Compare,
		LatestOnly: query.LatestOnly,
	}
	filter := request.BuildStockFilter(query.StockFilterQuery)
	if !trends.IsPerformanceQueryValid(info) {
		response.BadRequest(ctx)
		return
	}
	trends.SetQueryMeta(&info)
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	snapshots := database.GetFilteredSnapshots(db, user, filter, info.Meta.PermitLimited)
	groupedInfo, timePeriods, clickThrough := trends.ProcessSnapshots(snapshots, info)
	result := trends.BuildSummary(groupedInfo, info, timePeriods, clickThrough)
	ctx.JSON(http.StatusOK, result)
}

func RegisterTrendsRoutes(router *gin.RouterGroup) {
	router.GET("/trends", middleware.Authenticate("AccessPermissions"), Trends)
	// TODO: remove. This remains for now for backwards compatibility.
	router.GET("/performance", middleware.Authenticate("AccessPermissions"), Trends)

}
