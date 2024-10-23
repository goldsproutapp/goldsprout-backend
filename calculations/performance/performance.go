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
	}
	timePeriods := timeListGetters[info.TimeKey](timeCategories.Items())
	focusTime := util.Map(timePeriods, timeFocus[info.TimeKey])
	return groups, append(timePeriods, info.Meta.SummaryLabel), focusTime
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
	yearStartMap := map[string][]models.StockSnapshot{}
	latestMap := map[string]models.StockSnapshot{}
	snapshotMap := map[uint]map[time.Time][]models.StockSnapshot{}
	now := time.Now()
    // TODO: this should be configurable for l10n
	yearStart := time.Date(now.Year(), 4, 6, 0, 0, 0, 0, time.UTC)
	if now.Month() < 4 || now.Month() == 4 && now.Day() < 6 {
		yearStart = yearStart.AddDate(-1, 0, 0)
	}
	for _, snapshot := range snapshots {
		key := snapshot.Key()
		if !util.ContainsKey(latestMap, key) ||
			latestMap[key].Date.Compare(snapshot.Date) == -1 {
			latestMap[key] = snapshot
		}
		if snapshot.Date.After(yearStart) {
			if util.ContainsKey(yearStartMap, key) {
				yearStartMap[key] = append(yearStartMap[key], snapshot)
			} else {
				yearStartMap[key] = []models.StockSnapshot{snapshot}
			}
		}
		if !util.ContainsKey(snapshotMap, snapshot.AccountID) {
			snapshotMap[snapshot.AccountID] = map[time.Time][]models.StockSnapshot{}
		}
		if !util.ContainsKey(snapshotMap[snapshot.AccountID], snapshot.Date) {
			snapshotMap[snapshot.AccountID][snapshot.Date] = []models.StockSnapshot{snapshot}
		} else {
			snapshotMap[snapshot.AccountID][snapshot.Date] = append(snapshotMap[snapshot.AccountID][snapshot.Date], snapshot)
		}
	}
	snapshotMapMerged := map[time.Time][]models.StockSnapshot{}
	for accountID, m := range snapshotMap {
	dateLoop:
		for date, snapshots := range m {
			if util.ContainsKey(snapshotMapMerged, date) {
				continue dateLoop
			}
			allSnapshots := snapshots
			for otherAccount, otherMap := range snapshotMap {
				if accountID == otherAccount {
					continue
				}
				closest := time.Unix(0, 0)
				earliest := time.Now()
				var closestSnapshots []models.StockSnapshot
				for t, otherSnapshots := range otherMap {
					if t.Before(earliest) {
						earliest = t
					}
					if date.Sub(t).Abs().Seconds() < date.Sub(closest).Abs().Seconds() {
						closest = t
						closestSnapshots = otherSnapshots
					}
				}
				if !earliest.After(date) {
					allSnapshots = append(allSnapshots, closestSnapshots...)
				}
			}
			snapshotMapMerged[date] = allSnapshots
		}
	}

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
