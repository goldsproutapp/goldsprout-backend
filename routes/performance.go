package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/auth"
	"github.com/goldsproutapp/goldsprout-backend/calculations/performance"
	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/middleware"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/request"
	"github.com/goldsproutapp/goldsprout-backend/util"
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
    performance.SetQueryMeta(&info)
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	snapshots := database.FetchPerformanceData(db, user, filter, info.Meta.PermitLimited)
	groupedInfo, timePeriods, clickThrough := performance.ProcessSnapshots(snapshots, info)
	result := performance.BuildSummary(groupedInfo, info, timePeriods, clickThrough)
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
	request.OK(ctx, gin.H{"value": value, "performance": perf})
}

func PortfolioPerformance(ctx *gin.Context) {
	user := middleware.GetUser(ctx)
	db := middleware.GetDB(ctx)
	snapshots := database.GetSnapshots([]uint{user.ID}, []uint{}, db)
	info := performance.GeneratePerformanceGraphInfo(snapshots)
	request.OK(ctx, info)
}

func AccountPerformance(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
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
	account, err := database.GetAccount(db, uint(id))
	if err != nil {
		request.NotFound(ctx)
		return
	}
	user := middleware.GetUser(ctx)
	if !auth.HasAccessPerm(user, account.UserID, true, false, false) {
		request.Forbidden(ctx)
		return
	}
	snapshots := database.GetAccountSnapshots(uint(id), db)
	info := performance.GeneratePerformanceGraphInfo(snapshots)
	request.OK(ctx, info)
}

func RegisterPerformanceRoutes(router *gin.RouterGroup) {
	router.GET("/performance", middleware.Authenticate("AccessPermissions"), Perfomance)
	router.GET("/stockperformance", middleware.Authenticate("AccessPermissions"), StockPerformance)
	router.GET("/portfolioperformance", middleware.Authenticate(), PortfolioPerformance)
	router.GET("/accountperformance", middleware.Authenticate("AccessPermissions"), AccountPerformance)
}
