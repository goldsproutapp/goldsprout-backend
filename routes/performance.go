package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker-backend/calculations/performance"
	"github.com/patrickjonesuk/investment-tracker-backend/database"
	"github.com/patrickjonesuk/investment-tracker-backend/middleware"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/request"
)


func Perfomance(ctx *gin.Context) {
    var query models.PerformanceRequestQuery
    err := ctx.BindQuery(&query)
    if err != nil {
        request.BadRequest(ctx)
        return
    }
    info := models.PerformanceQueryInfo{
        TargetKey: query.Of,
        AgainstKey: query.For,
        TimeKey: query.Over,
        MetricKey: query.Compare,
    }
    if !calculations.IsPerformanceQueryValid(info) {
        request.BadRequest(ctx)
        return
    }
    db := middleware.GetDB(ctx)
    user := middleware.GetUser(ctx)
    snapshots := database.FetchPerformanceData(db, user)
    groupedInfo, timePeriods := calculations.ProcessSnapshots(snapshots, info)
    result := calculations.BuildSummary(groupedInfo, info, timePeriods)
    ctx.JSON(http.StatusOK, result)
}


func RegisterPerformanceRoutes(router *gin.RouterGroup) {
    router.GET("/performance", middleware.Authenticate("AccessPermissions"), Perfomance)
}

