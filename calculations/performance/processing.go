package performance

import (
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/goldsproutapp/goldsprout-backend/constants"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"github.com/shopspring/decimal"
)

var singlePropGetters = map[string]func(models.StockSnapshot) string{
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
var multiPropGetters = map[string]func(models.StockSnapshot) []string{
	"class": func(snapshot models.StockSnapshot) []string {
		return util.MapKeys(snapshot.Stock.ClassCompositionMap)
	},
}
var propGetters = map[string]func(models.StockSnapshot) []string{}

func SplitSnapshotValueByPercentage(snapshot models.StockSnapshot, percentage decimal.Decimal) models.StockSnapshot {
	pct := percentage.Div(decimal.NewFromInt(100))
	return models.StockSnapshot{
		ID:                     snapshot.ID,
		User:                   snapshot.User,
		UserID:                 snapshot.UserID,
		Account:                snapshot.Account,
		AccountID:              snapshot.AccountID,
		Date:                   snapshot.Date,
		Stock:                  snapshot.Stock,
		StockID:                snapshot.StockID,
		Units:                  snapshot.Units.Mul(pct),
		Price:                  snapshot.Price,
		Cost:                   snapshot.Cost.Mul(pct),
		Value:                  snapshot.Value.Mul(pct),
		ChangeToDate:           snapshot.ChangeToDate.Mul(pct),
		ChangeSinceLast:        snapshot.ChangeSinceLast.Mul(pct),
		NormalisedPerformance:  snapshot.NormalisedPerformance,
		TransactionAttribution: snapshot.TransactionAttribution,
	}
}

var compositionSplitters = map[string]func(models.StockSnapshot, string) models.StockSnapshot{
	"class": func(snapshot models.StockSnapshot, key string) models.StockSnapshot {
		s := SplitSnapshotValueByPercentage(snapshot, snapshot.Stock.ClassCompositionMap[key])
		s.Stock.ClassCompositionMap = map[string]decimal.Decimal{key: decimal.NewFromInt(100)}
		return s
	},
}

var SingleTargets = util.MapKeys(singlePropGetters)
var MultiTargets = util.MapKeys(multiPropGetters)
var Targets = append(SingleTargets, MultiTargets...)

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

func GetKeysFromSnapshot(snapshot models.StockSnapshot, key string) []string {
	if util.ContainsKey(singlePropGetters, key) {
		return util.Only(singlePropGetters[key](snapshot))
	} else {
		return multiPropGetters[key](snapshot)
	}
}

func GetContributionForCategory(snapshot models.StockSnapshot, key string, category string) models.StockSnapshot {
	if len(GetKeysFromSnapshot(snapshot, key)) == 1 {
		return snapshot
	}
	return compositionSplitters[key](snapshot, category)
}

func groupHasSnapshot(key string, value string, snapshot models.StockSnapshot) bool {
	return slices.Contains(GetKeysFromSnapshot(snapshot, key), value)
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
		Accounts:  util.Split(query.FilterAccounts, ","),
		LowerDate: ignoreBefore,
		UpperDate: ignoreAfter,
	}
	return filter
}
