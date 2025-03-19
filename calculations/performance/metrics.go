package performance

import (
	"time"

	"github.com/goldsproutapp/goldsprout-backend/constants"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"github.com/shopspring/decimal"
)

// Average monthly performance across all holdings in each period.
func PerformanceMetric(timeMap map[string][]models.StockSnapshot,
) map[string]decimal.Decimal {
	total := decimal.NewFromInt(0)
	items := map[string]decimal.Decimal{}
	for timePeriod, snapshots := range timeMap {
		if len(snapshots) == 0 {
			items[timePeriod] = decimal.NewFromInt(0)
			continue
		}
		subtotal := decimal.NewFromInt(0)
		for _, snapshot := range snapshots {
			subtotal = subtotal.Add(snapshot.NormalisedPerformance)
		}
		avg := subtotal.Div(decimal.NewFromInt(int64(len(snapshots)))).
			Truncate(constants.PERFORMANCE_DECIMAL_DIGITS)
		items[timePeriod] = avg
		total = total.Add(avg)
	}
	items[MetricMeta["performance"].SummaryLabel] = total.Div(decimal.NewFromInt(int64(len(timeMap)))).Truncate(constants.PERFORMANCE_DECIMAL_DIGITS)
	return items
}

// Average monthly performance across all holdings in each period, weighted by value.
func WeightedPerformanceMetric(timeMap map[string][]models.StockSnapshot,
) map[string]decimal.Decimal {
	items := map[string]decimal.Decimal{}
	total := decimal.NewFromInt(0)
	totalWeights := decimal.NewFromInt(0)
	for timePeriod, snapshots := range timeMap {
		if len(snapshots) == 0 {
			items[timePeriod] = decimal.NewFromInt(0)
			continue
		}
		subtotal := decimal.NewFromInt(0)
		weights := decimal.NewFromInt(0)
		for _, snapshot := range snapshots {
			weights = weights.Add(snapshot.Value)
			subtotal = subtotal.Add(snapshot.NormalisedPerformance.Mul(snapshot.Value))
		}
		items[timePeriod] = subtotal.Div(weights).Truncate(constants.PERFORMANCE_DECIMAL_DIGITS)
		total = total.Add(subtotal)
		totalWeights = totalWeights.Add(weights)

	}
	items[MetricMeta["weighted_performance"].SummaryLabel] = total.Div(totalWeights).Truncate(constants.PERFORMANCE_DECIMAL_DIGITS)
	return items
}

// Approximate growth across all holdings over the time period.
func GrowthMetric(timeMap map[string][]models.StockSnapshot) map[string]decimal.Decimal {
	items := map[string]decimal.Decimal{}

	oldestTotal := map[string]models.StockSnapshot{}
	newestTotal := map[string]models.StockSnapshot{}
	gainsTotal := map[string]decimal.Decimal{}
	for timePeriod, snapshots := range timeMap {
		if len(snapshots) == 0 {
			items[timePeriod] = decimal.NewFromInt(0)
			continue
		}
		oldest := map[string]models.StockSnapshot{}
		newest := map[string]models.StockSnapshot{}
		valueSum := decimal.NewFromInt(0)
		gains := map[string]decimal.Decimal{}
		for _, snapshot := range snapshots {
			key := snapshot.Key()
			if !util.ContainsKey(oldest, key) || snapshot.Date.Before(oldest[key].Date) {
				oldest[key] = snapshot
			}
			if !util.ContainsKey(newest, key) || snapshot.Date.After(newest[key].Date) {
				newest[key] = snapshot
			}
			if !util.ContainsKey(oldestTotal, key) || snapshot.Date.Before(oldestTotal[key].Date) {
				oldestTotal[key] = snapshot
			}
			if !util.ContainsKey(newestTotal, key) || snapshot.Date.After(newestTotal[key].Date) {
				newestTotal[key] = snapshot
			}
			if !util.ContainsKey(gains, key) {
				gains[key] = decimal.NewFromInt(0)
			}
			if !util.ContainsKey(gainsTotal, key) {
				gainsTotal[key] = decimal.NewFromInt(0)
			}
			gains[key] = gains[key].Add(snapshot.ChangeSinceLast)
			gainsTotal[key] = gainsTotal[key].Add(snapshot.ChangeSinceLast)
		}
		gainPerPerfTotal := decimal.NewFromInt(0)
		totalGain := decimal.NewFromInt(0)
		for key, snapshot := range newest {
			valueSum = valueSum.Sub(snapshot.Value)
			perf := snapshot.Price.Div(oldest[key].Price).Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100))
			if !perf.Equal(decimal.NewFromInt(0)) {
				totalGain = totalGain.Add(gains[key])
				gainPerPerf := gains[key].Div(perf)
				gainPerPerfTotal = gainPerPerfTotal.Add(gainPerPerf)
			}
		}
		if gainPerPerfTotal.Equal(decimal.NewFromInt(0)) {
			items[timePeriod] = decimal.NewFromInt(0)
		} else {
			items[timePeriod] = totalGain.Div(gainPerPerfTotal).Truncate(2)
		}
	}
	gainPerPerfTotal := decimal.NewFromInt(0)
	totalGain := decimal.NewFromInt(0)
	for key, snapshot := range newestTotal {
		perf := snapshot.Price.Div(oldestTotal[key].Price).Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100))
		if !perf.Equal(decimal.NewFromInt(0)) {
			totalGain = totalGain.Add(gainsTotal[key])
			gainPerPerf := gainsTotal[key].Div(perf)
			gainPerPerfTotal = gainPerPerfTotal.Add(gainPerPerf)
		}
	}
	var out decimal.Decimal
	if gainPerPerfTotal.Equal(decimal.NewFromInt(0)) {
		out = decimal.NewFromInt(0)
	} else {
		out = totalGain.Div(gainPerPerfTotal).Truncate(2)
	}
	items[MetricMeta["growth"].SummaryLabel] = out

	return items
}

// Total value of all holdings at the end of each time period
func HoldingsMetric(timeMap map[string][]models.StockSnapshot,
) map[string]decimal.Decimal {

	items := map[string]decimal.Decimal{}
	latestDateTotal := time.Unix(0, 0)
	latestTimePeriod := ""
	for timePeriod, snapshots := range timeMap {
		snapshotMap := map[string]models.StockSnapshot{}
		latestForPeriod := map[uint]time.Time{}
		for _, snapshot := range snapshots {
            key := snapshot.Key()
			latest, existsLatest := snapshotMap[key]
			if !existsLatest || snapshot.Date.Compare(latest.Date) == 1 {
				snapshotMap[key] = snapshot
			}
			if !util.ContainsKey(latestForPeriod, snapshot.AccountID) || snapshot.Date.Compare(latestForPeriod[snapshot.AccountID]) == 1 {
				latestForPeriod[snapshot.AccountID] = snapshot.Date
			}
			if snapshot.Date.Compare(latestDateTotal) == 1 {
				latestDateTotal = snapshot.Date
				latestTimePeriod = timePeriod
			}
		}
		timeTotal := decimal.NewFromInt(0)
		for _, snapshot := range snapshotMap {
			if snapshot.Date.Sub(latestForPeriod[snapshot.AccountID]).Abs().Minutes() < 60 { // If you're doing subsequent imports less than an hour apart then it's your fault that this doesn't work for you.
				timeTotal = timeTotal.Add(snapshot.Value)
			}
		}
		items[timePeriod] = timeTotal
	}
	items[MetricMeta["holdings"].SummaryLabel] = items[latestTimePeriod]
	return items
}

func GainsMetric(timeMap map[string][]models.StockSnapshot,
) map[string]decimal.Decimal {

	items := map[string]decimal.Decimal{}
	total := decimal.NewFromInt(0)
	for timePeriod, snapshots := range timeMap {
		timeTotal := decimal.NewFromInt(0)
		for _, snapshot := range snapshots {
			timeTotal = timeTotal.Add(snapshot.ChangeSinceLast)
			total = total.Add(snapshot.ChangeSinceLast)
		}
		items[timePeriod] = timeTotal
	}
	items[MetricMeta["gains"].SummaryLabel] = total
	return items
}

var metricsMap = map[string]func(
	timeMap map[string][]models.StockSnapshot,
) map[string]decimal.Decimal{

	"performance": PerformanceMetric,

	"weighted_performance": WeightedPerformanceMetric,

	"growth": GrowthMetric,

	"holdings": HoldingsMetric,

	"gains": GainsMetric,
}

var MetricMeta = map[string]models.PerformanceMetricMeta{
	"performance": models.PerformanceMetricMeta{
		PermitLimited: true,
		SummaryLabel:  "Average",
	},
	"weighted_performance": models.PerformanceMetricMeta{
		PermitLimited: true,
		SummaryLabel:  "Average",
	},
	"growth": models.PerformanceMetricMeta{
		PermitLimited: true,
		SummaryLabel:  "Total",
	},
	"holdings": models.PerformanceMetricMeta{
		PermitLimited: false,
		SummaryLabel:  "Latest",
	},
	"gains": models.PerformanceMetricMeta{
		PermitLimited: false,
		SummaryLabel:  "Total",
	},
}

var Metrics = util.MapKeys(metricsMap)
