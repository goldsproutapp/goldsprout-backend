package performance

import (
	"slices"

	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/util"
	"github.com/shopspring/decimal"
)

func IsPerformanceQueryValid(p models.PerformanceQueryInfo) bool {
	return slices.Contains(Targets, p.TargetKey) &&
		slices.Contains(Targets, p.AgainstKey) &&
		slices.Contains(Metrics, p.MetricKey) &&
		slices.Contains(Times, p.TimeKey)
}

func ProcessSnapshots(snapshots []models.StockSnapshot, info models.PerformanceQueryInfo) (models.PerformanceMap, []string) {
	groups := models.PerformanceMap{}
	timeCategories := util.NewOrderedSet[string]()
	for _, snapshot := range snapshots {
		target := GetKeyFromSnapshot(snapshot, info.TargetKey)
		timeCategory := getTimeCategoryFromSnapshot(snapshot, info.TimeKey)
		timeCategories.Add(timeCategory)
		against := GetKeyFromSnapshot(snapshot, info.AgainstKey)

		addSnapshotToMap(&groups, snapshot, target, against, timeCategory)
	}
	return groups, append(timeListGetters[info.TimeKey](timeCategories.Items()), SummaryLabels[info.MetricKey])
}

func BuildSummary(perfMap models.PerformanceMap, info models.PerformanceQueryInfo, timePeriods []string) models.PerformanceResponse {
	res := models.PerformanceResponse{
		TimePeriods: timePeriods,
		Data:        map[string]models.CategoryPerformance{},
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
			totalMap[timePeriod] = []models.StockSnapshot{}
			for _, timeMap := range groups {
				snapshots := timeMap[timePeriod]
				totalMap[timePeriod] = append(totalMap[timePeriod], snapshots...)
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
