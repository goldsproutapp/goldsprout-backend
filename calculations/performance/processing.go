package performance

import (
	"sort"
	"strconv"
	"time"

	"github.com/patrickjonesuk/investment-tracker-backend/constants"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/util"
)

var propGetters = map[string]func(models.StockSnapshot) string{
	"person": func(snapshot models.StockSnapshot) string {
		/*
			 TODO: ideally categorisation will be by id and display by name,
						but currently the same value is required for both.
		*/
		return snapshot.User.FirstName + " " + snapshot.User.LastName
	},
	"provider": func(snapshot models.StockSnapshot) string {
		return snapshot.Stock.Provider.Name
	},
	"account": func(snapshot models.StockSnapshot) string {
		return snapshot.Account.Name
	},
	"sector": func(snapshot models.StockSnapshot) string {
		return snapshot.Stock.Sector
	},
	"region": func(snapshot models.StockSnapshot) string {
		return snapshot.Stock.Region
	},
	"stock": func(snapshot models.StockSnapshot) string {
		return snapshot.Stock.Name
	},
	"all": func(_ models.StockSnapshot) string {
		return ""
	},
}

var Targets = util.MapKeys(propGetters)

var timeGetters = map[string]func(models.StockSnapshot) string{
	"years": func(snapshot models.StockSnapshot) string {
		return strconv.FormatInt(int64(snapshot.Date.Year()), 10)
	},
	"months": func(snapshot models.StockSnapshot) string {
		return snapshot.Date.Month().String()
	},
}
var Times = util.MapKeys(timeGetters)

var timeListGetters = map[string]func([]string) []string{
	"years": func(years []string) []string {
		sort.Slice(years, func(a, b int) bool {
			errList := []error{}
			return util.ParseUint(years[a], &errList) < util.ParseUint(years[b], &errList)
		})
		return years
	},
	"months": func(moonths []string) []string {
		return constants.MONTHS
	},
}

var timeFocus = map[string]func(string) []string{
	"years": func(year string) []string {
		yearNum := util.ParseIntOrDefault(year, -1)
		start := time.Date(yearNum, time.January, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(yearNum, time.December, 31, 23, 0, 0, 0, time.UTC)
		return []string{"months", strconv.FormatInt(start.Unix(), 10), strconv.FormatInt(end.Unix(), 10)}
	},
	"months": func(month string) []string {
		return []string{}
	},
}

func GetKeyFromSnapshot(snapshot models.StockSnapshot, key string) string {
	return propGetters[key](snapshot)
}

func groupHasSnapshot(key string, value string, snapshot models.StockSnapshot) bool {
	return GetKeyFromSnapshot(snapshot, key) == value
}

func getTimeCategoryFromSnapshot(snapshot models.StockSnapshot, timeKey string) string {
	return timeGetters[timeKey](snapshot)
}

func BuildStockFilter(query models.StockFilterQuery) models.StockFilter {
	ignoreBefore := time.Unix(int64(util.ParseIntOrDefault(query.FilterIgnoreBefore, 0)), 0)
	ignoreAfter := time.Unix(int64(util.ParseIntOrDefault(query.FilterIgnoreAfter, 0)), 0)
	filter := models.StockFilter{
		Regions:   util.Split(query.FilterRegions, ","),
		Providers: util.UintArray(query.FilterProviders),
		Users:     util.UintArray(query.FilterUsers),
		LowerDate: ignoreBefore,
		UpperDate: ignoreAfter,
	}
	return filter
}
