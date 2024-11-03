package performance

import (
	"slices"
	"time"

	"github.com/goldsproutapp/goldsprout-backend/calculations"
	"github.com/goldsproutapp/goldsprout-backend/constants"
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

func SetQueryMeta(p *models.PerformanceQueryInfo) {
	p.Meta = MetricMeta[p.MetricKey]
	p.MetricFunction = metricsMap[p.MetricKey]
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
		if info.GenerateSummary() {
			addSnapshotToMap(&groups, snapshot, constants.TRENDS_SUMMARY, "", timeCategory)
		}
	}
	timePeriods := timeListGetters[info.TimeKey](timeCategories.Items())
	focusTime := util.Map(timePeriods, timeFocus[info.TimeKey])
	return groups, append(timePeriods, info.Meta.SummaryLabel), focusTime
}

func BuildSummary(perfMap models.PerformanceMap, info models.PerformanceQueryInfo, timePeriods []string, timeFocus [][]string) models.PerformanceResponse {
	var summary string
	if info.GenerateSummary() {
		summary = constants.TRENDS_SUMMARY
	} else {
		summary = ""
	}
	res := models.PerformanceResponse{
		TimePeriods: timePeriods,
		Data:        map[string]models.CategoryPerformance{},
		TimeFocus:   timeFocus,
		SummaryRow:  summary,
	}
	for target, groups := range perfMap {
		category := models.CategoryPerformance{
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

	snapshotMapMerged, yearStartMap := calculations.CreateMergedSnapshotMap(snapshots)

	valueOut := map[time.Time]decimal.Decimal{}
	costOut := map[time.Time]decimal.Decimal{}
	perfOut := map[time.Time]decimal.Decimal{}
	for time, snapshotList := range snapshotMapMerged {
		valueOut[time] = decimal.NewFromInt(0)
		costOut[time] = decimal.NewFromInt(0)
		perfOut[time] = decimal.NewFromInt(0)
		counted := map[string]models.StockSnapshot{}
		for _, snapshot := range snapshotList {
			p := snapshot.NormalisedPerformance.Mul(snapshot.Value)
			if util.ContainsKey(counted, snapshot.Key()) {
				if counted[snapshot.Key()].Date.After(snapshot.Date) {
					continue
				} else {
					other := counted[snapshot.Key()]
					valueOut[time] = valueOut[time].Sub(other.Value)
					costOut[time] = costOut[time].Sub(other.Cost)
					perfOut[time] = perfOut[time].Sub(other.NormalisedPerformance.Mul(other.Value))
				}
			}
			valueOut[time] = valueOut[time].Add(snapshot.Value)
			costOut[time] = costOut[time].Add(snapshot.Cost)
			perfOut[time] = perfOut[time].Add(p)
			counted[snapshot.Key()] = snapshot
		}
		perfOut[time] = perfOut[time].Div(valueOut[time]).Truncate(2)
	}
	totalGPP := decimal.NewFromInt(0)
	totalGain := decimal.NewFromInt(0)
	for _, snapshots := range yearStartMap {
		perf := snapshots[len(snapshots)-1].Price.Div(snapshots[0].Price).Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100))
		if perf.Equal(decimal.NewFromInt(0)) {
			continue
		}
		gain := decimal.Sum(decimal.NewFromInt(0), util.Map(snapshots, func(s models.StockSnapshot) decimal.Decimal {
			return s.ChangeSinceLast
		})...)
		gainPerPerf := gain.Div(perf)
		totalGain = totalGain.Add(gain)
		totalGPP = totalGPP.Add(gainPerPerf)
	}
	var ytd decimal.Decimal
	zero := decimal.NewFromInt(0)
	if totalGPP.Equal(zero) {
		ytd = zero
	} else {
		ytd = totalGain.Div(totalGPP).Truncate(2)
	}
	return models.PerformanceGraphInfo{
		Performance: perfOut,
		Value:       valueOut,
		Cost:        costOut,
		YearToDate:  ytd,
	}
}
