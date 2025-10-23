package metrics

import (
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"github.com/shopspring/decimal"
)

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
	items[GetMetricMetaByName("growth").SummaryLabel] = out

	return items
}

