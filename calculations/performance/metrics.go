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

var metricsMap = map[string]func(
	timeMap map[string][]models.StockSnapshot,
) map[string]decimal.Decimal{

	"performance": PerformanceMetric,

	"weighted_performance": WeightedPerformanceMetric,

	"holdings": HoldingsMetric,
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
	"holdings": models.PerformanceMetricMeta{
		PermitLimited: false,
		SummaryLabel:  "Latest",
	},
}

var Metrics = util.MapKeys(metricsMap)
