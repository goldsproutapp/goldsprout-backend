package performance

import (
	"slices"
	"time"

	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"github.com/shopspring/decimal"
)

func IsPerformanceQueryValid(p models.PerformanceQueryInfo) bool {
	return slices.Contains(Targets, p.TargetKey) &&
		slices.Contains(Targets, p.AgainstKey) &&
		slices.Contains(Metrics, p.MetricKey) &&
		slices.Contains(Times, p.TimeKey)
}

func ProcessSnapshots(snapshots []models.StockSnapshot, info models.PerformanceQueryInfo) (models.PerformanceMap, []string, [][]string) {
	groups := models.PerformanceMap{}
	timeCategories := util.NewOrderedSet[string]()
	for _, snapshot := range snapshots {
		target := GetKeyFromSnapshot(snapshot, info.TargetKey)
		timeCategory := getTimeCategoryFromSnapshot(snapshot, info.TimeKey)
		timeCategories.Add(timeCategory)
		against := GetKeyFromSnapshot(snapshot, info.AgainstKey)

		addSnapshotToMap(&groups, snapshot, target, against, timeCategory)
	}
	timePeriods := timeListGetters[info.TimeKey](timeCategories.Items())
	focusTime := util.Map(timePeriods, timeFocus[info.TimeKey])
	return groups, append(timePeriods, SummaryLabels[info.MetricKey]), focusTime
}

func BuildSummary(perfMap models.PerformanceMap, info models.PerformanceQueryInfo, timePeriods []string, timeFocus [][]string) models.PerformanceResponse {
	res := models.PerformanceResponse{
		TimePeriods: timePeriods,
		Data:        map[string]models.CategoryPerformance{},
		TimeFocus:   timeFocus,
	}
	for target, groups := range perfMap {
		category := models.CategoryPerformance{
			Totals: map[string]decimal.Decimal{},
			Items:  map[string]map[string]decimal.Decimal{},
		}
		for group, timeMap := range groups {
			items := metricsMap[info.MetricKey](timeMap)
			category.Items[group] = items
		}
		totalMap := map[string][]models.StockSnapshot{}
		for _, timePeriod := range timePeriods {
			periodSnapshots := []models.StockSnapshot{}
			for _, timeMap := range groups {
				snapshots := timeMap[timePeriod]
				periodSnapshots = append(periodSnapshots, snapshots...)
			}
			if len(periodSnapshots) > 0 {
				totalMap[timePeriod] = periodSnapshots
			}
		}
		totals := metricsMap[info.MetricKey](totalMap)
		category.Totals = totals
		res.Data[target] = category
	}

	return res
}

func addSnapshotToMap(m *models.PerformanceMap, snapshot models.StockSnapshot, a string, b string, c string) {
	_, ok := (*m)[a]
	if !ok {
		(*m)[a] = map[string]map[string][]models.StockSnapshot{}
	}
	_, ok = (*m)[a][b]
	if !ok {
		(*m)[a][b] = map[string][]models.StockSnapshot{}
	}
	_, ok = (*m)[a][b][c]
	if !ok {
		(*m)[a][b][c] = []models.StockSnapshot{}
	}
	(*m)[a][b][c] = append((*m)[a][b][c], snapshot)
}

func GeneratePerformanceGraphInfo(snapshots []models.StockSnapshot) models.PerformanceGraphInfo {
	perf := map[time.Time][]decimal.Decimal{}
	value := map[time.Time][]decimal.Decimal{}
	yearAgoMap := map[uint]models.StockSnapshot{}
	latestMap := map[uint]models.StockSnapshot{}
	for _, snapshot := range snapshots {
		if !util.ContainsKey(yearAgoMap, snapshot.StockID) &&
			time.Now().Sub(snapshot.Date).Hours() < (24*365) {
			yearAgoMap[snapshot.StockID] = snapshot
		}
		if !util.ContainsKey(latestMap, snapshot.StockID) ||
			latestMap[snapshot.StockID].Date.Compare(snapshot.Date) == -1 {
			latestMap[snapshot.StockID] = snapshot
		}

		if util.ContainsKey(perf, snapshot.Date) {
			perf[snapshot.Date] = append(perf[snapshot.Date], snapshot.NormalisedPerformance.Mul(snapshot.Value))
		} else {
			perf[snapshot.Date] = []decimal.Decimal{snapshot.NormalisedPerformance.Mul(snapshot.Value)}
		}
		if util.ContainsKey(value, snapshot.Date) {
			value[snapshot.Date] = append(value[snapshot.Date], snapshot.Value)
		} else {
			value[snapshot.Date] = []decimal.Decimal{snapshot.Value}
		}
	}
	valueOut := map[time.Time]decimal.Decimal{}
	for time, list := range value {
		valueOut[time] = decimal.Sum(list[0], list[1:]...).Truncate(2)
	}
	perfOut := map[time.Time]decimal.Decimal{}
	for time, list := range perf {
		perfOut[time] = decimal.Sum(list[0], list[1:]...).Div(valueOut[time]).Truncate(2)
	}
	old := decimal.NewFromInt(0)
	n := decimal.NewFromInt(0)
	for sid, snapshot := range yearAgoMap {
		if !util.ContainsKey(latestMap, sid) {
			continue
		}
		latest := latestMap[sid]
		old = old.Add(snapshot.Price.Mul(latest.Units))
		n = n.Add(latest.Price.Mul(latest.Units))
	}
	var ytd decimal.Decimal
	zero := decimal.NewFromInt(0)
	if n.Equal(zero) {
		ytd = zero
	} else {
		ytd = n.Sub(old).Div(n).Mul(decimal.NewFromInt(100)).Truncate(2)
	}
	return models.PerformanceGraphInfo{
		Performance: perfOut,
		Value:       valueOut,
		YearToDate:  ytd,
	}
}
