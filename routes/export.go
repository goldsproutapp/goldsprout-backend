package routes

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker-backend/constants"
	"github.com/patrickjonesuk/investment-tracker-backend/database"
	"github.com/patrickjonesuk/investment-tracker-backend/middleware"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/request"
	"gorm.io/gorm/clause"
)

var headings = []string{
	"date",
	"user",
	"provider",
	"stock_code",
	"stock_name",
	"region",
	"sector",
	"annual_fee",
	"units",
	"price",
	"cost",
	"value",
	"absolute_change",
	"normalised_performance",
}

func FormatCSV(snapshot models.StockSnapshot) string {
	fields := []string{
		snapshot.Date.Format(constants.ISO8601),
		snapshot.User.Name(),
		snapshot.Stock.Provider.Name,
		snapshot.Stock.StockCode,
		snapshot.Stock.Name,
		snapshot.Stock.Region,
		snapshot.Stock.Sector,
		strconv.FormatFloat(float64(snapshot.Stock.AnnualFee), 'f', 2, 64),
		snapshot.Units.String(),
		snapshot.Price.String(),
		snapshot.Cost.String(),
		snapshot.Value.String(),
		snapshot.ChangeToDate.String(),
		snapshot.NormalisedPerformance.String(),
	}
	return strings.Join(fields, ",")
}

func ExportToCSV(ctx *gin.Context) {
	user := middleware.GetUser(ctx)
	db := middleware.GetDB(ctx)
	snapshots := database.GetAllSnapshots(user, db, clause.Associations, "Stock.Provider")
	outputArr := make([]string, len(snapshots)+1)
	outputArr[0] = strings.Join(headings, ",")
	for i, snapshot := range snapshots {
		str := FormatCSV(snapshot)
		outputArr[i+1] = str
	}
	output := strings.Join(outputArr, "\n")
	request.FileOK(ctx, "export.csv", output)
}

func RegisterExportRoutes(router *gin.RouterGroup) {
	router.GET("/export/csv", middleware.Authenticate("AccessPermissions"), ExportToCSV)
}
