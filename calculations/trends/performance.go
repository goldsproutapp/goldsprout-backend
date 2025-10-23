package trends

import (
	"github.com/goldsproutapp/goldsprout-backend/lib/extraction"
	"github.com/goldsproutapp/goldsprout-backend/lib/extraction/times"
	"github.com/goldsproutapp/goldsprout-backend/constants"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"github.com/shopspring/decimal"
)

func ProcessSnapshots(snapshots []models.StockSnapshot, info PerformanceQueryInfo) (PerformanceMap, []string, [][]string) {
	groups := PerformanceMap{}
	timeCategories := util.NewOrderedSet[string]()
	for _, snapshot := range snapshots {
		targetValues := extraction.GetKeysFromSnapshot(snapshot, info.TargetKey)
		timeCategory := extraction.ExtractTimeFromSnapshot(times.PerformanceTimeExtractionSet(), info.TimeKey, snapshot)
		timeCategories.Add(timeCategory)
		againstValues := extraction.GetKeysFromSnapshot(snapshot, info.AgainstKey)

		for _, target := range targetValues {
			for _, against := range againstValues {
				s := extraction.GetContributionForCategory(extraction.GetContributionForCategory(snapshot, info.TargetKey, target), info.AgainstKey, against)
				addSnapshotToMap(&groups, s, target, against, timeCategory)
			}
		}
		if info.GenerateSummary() {
			addSnapshotToMap(&groups, snapshot, constants.TRENDS_SUMMARY, "", timeCategory)
		}
	}
	if info.GenerateSummary() && len(groups) == 2 {
		delete(groups, constants.TRENDS_SUMMARY)
	}
	timePeriods := extraction.ExtractTimeList(info.TimeKey, timeCategories.Items())
	focusTime := util.Map(timePeriods, extraction.GetTimeFocusFunction(info.TimeKey))
	return groups, append(timePeriods, info.Meta.SummaryLabel), focusTime
}

func BuildSummary(perfMap PerformanceMap, info PerformanceQueryInfo, timePeriods []string, timeFocus [][]string) PerformanceResponse {
	var summary string
	if info.GenerateSummary() && util.ContainsKey(perfMap, constants.TRENDS_SUMMARY) {
		summary = constants.TRENDS_SUMMARY
	} else {
		summary = ""
	}
	res := PerformanceResponse{
		TimePeriods: timePeriods,
		Data:        map[string]CategoryPerformance{},
		TimeFocus:   timeFocus,
		SummaryRow:  summary,
	}
	for target, groups := range perfMap {
		category := CategoryPerformance{
			Totals: map[string]decimal.Decimal{},
			Items:  map[string]map[string]decimal.Decimal{},
		}
		for group, timeMap := range groups {
			items := info.MetricFunction(timeMap)
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
		totals := info.MetricFunction(totalMap)
		category.Totals = totals
		res.Data[target] = category
	}

	return res
}

func addSnapshotToMap(m *PerformanceMap, snapshot models.StockSnapshot, a string, b string, c string) {
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
