package request

import (
	"time"

	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
)

func BuildStockFilter(query models.StockFilterQuery) database.StockFilter {
	ignoreBefore := time.Unix(int64(util.ParseIntOrDefault(query.FilterIgnoreBefore, 0)), 0)
	ignoreAfter := time.Unix(int64(util.ParseIntOrDefault(query.FilterIgnoreAfter, 0)), 0)
	filter := database.StockFilter{
		Regions:   util.Split(query.FilterRegions, ","),
		Providers: util.UintArray(query.FilterProviders),
		Users:     util.UintArray(query.FilterUsers),
		Accounts:  util.Split(query.FilterAccounts, ","),
		LowerDate: ignoreBefore,
		UpperDate: ignoreAfter,
	}
	return filter
}
