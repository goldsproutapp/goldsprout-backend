package routes

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker-backend/calculations/performance"
	"github.com/patrickjonesuk/investment-tracker-backend/database"
	"github.com/patrickjonesuk/investment-tracker-backend/middleware"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/request"
	"github.com/patrickjonesuk/investment-tracker-backend/util"
)

func UintArray(input string) []uint {
    output := []uint{}
    errList := []error{}
    for _, item := range Split(input, ",") {
        numErrors := len(errList)
        number := util.ParseUint(item, &errList)
        if len(errList) == numErrors {
            output = append(output, number)
        }
    }
    return output
}
func Split(input string, sep string) []string {
    if len(input) == 0 {
        return make([]string, 0)
    }
    return strings.Split(input, sep)
}


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
    filter := models.PerformanceFilter {
        Regions: Split(query.FilterRegions, ","),
        Providers: UintArray(query.FilterProviders),
        Users: UintArray(query.FilterUsers),
    }
    if !calculations.IsPerformanceQueryValid(info) {
        request.BadRequest(ctx)
        return
    }
    db := middleware.GetDB(ctx)
    user := middleware.GetUser(ctx)
    snapshots := database.FetchPerformanceData(db, user, filter)
    groupedInfo, timePeriods := calculations.ProcessSnapshots(snapshots, info)
    result := calculations.BuildSummary(groupedInfo, info, timePeriods)
    ctx.JSON(http.StatusOK, result)
}


func RegisterPerformanceRoutes(router *gin.RouterGroup) {
    router.GET("/performance", middleware.Authenticate("AccessPermissions"), Perfomance)
}

