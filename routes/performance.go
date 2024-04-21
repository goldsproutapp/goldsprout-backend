package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker-backend/calculations/performance"
	"github.com/patrickjonesuk/investment-tracker-backend/database"
	"github.com/patrickjonesuk/investment-tracker-backend/middleware"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/request"
	"github.com/patrickjonesuk/investment-tracker-backend/util"
	"github.com/shopspring/decimal"
)


func Perfomance(ctx *gin.Context) {
	var query models.PerformanceRequestQuery
	err := ctx.BindQuery(&query)
	if err != nil {
		request.BadRequest(ctx)
		return
	}
	info := models.PerformanceQueryInfo{
		TargetKey:  query.Of,
		AgainstKey: query.For,
		TimeKey:    query.Over,
		MetricKey:  query.Compare,
	}
	filter := performance.BuildStockFilter(query.StockFilterQuery)
	if !performance.IsPerformanceQueryValid(info) {
		request.BadRequest(ctx)
		return
	}
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	snapshots := database.FetchPerformanceData(db, user, filter)
	groupedInfo, timePeriods := performance.ProcessSnapshots(snapshots, info)
	result := performance.BuildSummary(groupedInfo, info, timePeriods)
	ctx.JSON(http.StatusOK, result)
}

func StockPerformance(ctx *gin.Context) {
	idstr, exists := ctx.GetQuery("id")
	if !exists {
		request.BadRequest(ctx)
		return
	}
	id := util.ParseIntOrDefault(idstr, -1)
	if id == -1 {
		request.BadRequest(ctx)
		return
	}
	db := middleware.GetDB(ctx)

	var snapshots []models.StockSnapshot
	db.Model(&models.StockSnapshot{}).
		// NOTE: this allows all users to see performance data from all other users.
		// This seems reasonable as it *shouldn't* be private in any way
		Where("stock_id = ?", id).
		Find(&snapshots)

	perf := map[time.Time]decimal.Decimal{}
	value := map[time.Time]decimal.Decimal{}
	for _, snapshot := range snapshots {
		if !util.ContainsKey(value, snapshot.Date) {
			value[snapshot.Date] = snapshot.Price
		}
		if !util.ContainsKey(perf, snapshot.Date) {
			perf[snapshot.Date] = snapshot.NormalisedPerformance
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"value": value, "performance": perf})
}
func PortfolioPerformance(ctx *gin.Context) {
	user := middleware.GetUser(ctx)
	db := middleware.GetDB(ctx)
	snapshots := database.GetAllSnapshots(user, db)
	perf := map[time.Time][]decimal.Decimal{}
    value := map[time.Time][]decimal.Decimal{}
	for _, snapshot := range snapshots {
		if util.ContainsKey(perf, snapshot.Date) {
			perf[snapshot.Date] = append(perf[snapshot.Date], snapshot.NormalisedPerformance)
		} else {
			perf[snapshot.Date] = []decimal.Decimal{snapshot.NormalisedPerformance}
		}
		if util.ContainsKey(value, snapshot.Date) {
			value[snapshot.Date] = append(value[snapshot.Date], snapshot.Value)
		} else {
            value[snapshot.Date] = []decimal.Decimal{snapshot.Value}
		}
	}
	perfOut := map[time.Time]decimal.Decimal{}
	for time, list := range perf {
		perfOut[time] = decimal.Avg(list[0], list[1:]...).Truncate(2)
	}
	valueOut := map[time.Time]decimal.Decimal{}
	for time, list := range value {
		valueOut[time] = decimal.Sum(list[0], list[1:]...).Truncate(2)
	}
    ctx.JSON(http.StatusOK, gin.H{"performance": perfOut, "value": valueOut})
}

func RegisterPerformanceRoutes(router *gin.RouterGroup) {
	router.GET("/performance", middleware.Authenticate("AccessPermissions"), Perfomance)
	router.GET("/stockperformance", middleware.Authenticate("AccessPermissions"), StockPerformance)
	router.GET("/portfolioperformance", middleware.Authenticate(), PortfolioPerformance)
}
