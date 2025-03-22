package routes

import (

	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/calculations/performance"
	"github.com/goldsproutapp/goldsprout-backend/calculations/reports"
	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/middleware"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/request"
)

func Report(ctx *gin.Context) {
	var query models.ReportRequestQuery
	err := ctx.BindQuery(&query)
	if err != nil || !reports.IsReportQueryValid(query) {
		request.BadRequest(ctx)
		return
	}
	filter := performance.BuildStockFilter(query.StockFilterQuery)
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	snapshots := database.FetchPerformanceData(db, user, filter, false)
	split, times := reports.SplitSnapshots(query.Period, snapshots)
	reportMap := map[string]models.Report{}
	lowestDate := filter.LowerDate
	if len(times) > 0 && len(split[times[0]]) > 0 {
		lowestDate = split[times[0]][0].Date
	}
	for p, s := range split {
		if len(s) == 0 {
			reportMap[p] = models.Report{
				Transactions: []models.ReportTransaction{}, // return empty array instead of `null`
			}
			continue
		}
		t := lowestDate
		if p != "Total" {
			t = reports.PrevPeriod[query.Period](s[0].Date)
		}
		aggregated := reports.AggregateSnapshots(db, t, s)
		report := reports.GenerateReport(aggregated)
		reportMap[p] = report
	}

	request.OK(ctx, gin.H{"periods": times, "report": reportMap})
}

func RegisterReportRoutes(router *gin.RouterGroup) {
	router.GET("/report", middleware.Authenticate("AccessPermissions"), Report)
}
