package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/calculations/reports"
	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/middleware"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/request"
	"github.com/goldsproutapp/goldsprout-backend/request/response"
)

func Report(ctx *gin.Context) {
	var query models.ReportRequestQuery
	err := ctx.BindQuery(&query)
	if err != nil || !reports.IsReportQueryValid(query) {
		response.BadRequest(ctx)
		return
	}
	filter := request.BuildStockFilter(query.StockFilterQuery)
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	snapshots := database.GetFilteredSnapshots(db, user, filter, false)
	times, reportMap := reports.CalculateReport(db, filter, query, snapshots)

	response.OK(ctx, gin.H{"periods": times, "report": reportMap})
}

func RegisterReportRoutes(router *gin.RouterGroup) {
	router.GET("/report", middleware.Authenticate("AccessPermissions"), Report)
}
