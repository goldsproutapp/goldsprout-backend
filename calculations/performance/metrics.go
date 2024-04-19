package calculations

import (
	"fmt"
	"time"

	"github.com/patrickjonesuk/investment-tracker-backend/constants"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/util"
	"github.com/shopspring/decimal"
)

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
	items["Total"] = total.Div(decimal.NewFromInt(int64(len(timeMap)))).Truncate(constants.PERFORMANCE_DECIMAL_DIGITS)
	return items
}

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
	items["Total"] = total.Div(totalWeights).Truncate(constants.PERFORMANCE_DECIMAL_DIGITS)
	return items
}

func HoldingsMetric(timeMap map[string][]models.StockSnapshot,
) map[string]decimal.Decimal {

	items := map[string]decimal.Decimal{}
    latestDateTotal := time.Unix(0, 0)
    latestTimePeriod := ""
	for timePeriod, snapshots := range timeMap {
		dateMap := map[string]time.Time{}
		valueMap := map[string]decimal.Decimal{}
		for _, snapshot := range snapshots {
            key := fmt.Sprintf("%d:%d", snapshot.StockID, snapshot.UserID)
			latest, existsLatest := dateMap[key]
			if !existsLatest || snapshot.Date.Compare(latest) == 1 {
				dateMap[key] = snapshot.Date
				valueMap[key] = snapshot.Value
			}
            if snapshot.Date.Compare(latestDateTotal) == 1 {
                latestDateTotal = snapshot.Date
                latestTimePeriod = timePeriod
            }
		}
		timeTotal := decimal.NewFromInt(0)
		for _, value := range valueMap {
			timeTotal = timeTotal.Add(value)
		}
		items[timePeriod] = timeTotal
	}
    items["Total"] = items[latestTimePeriod]
	return items
}

var metricsMap = map[string]func(
	timeMap map[string][]models.StockSnapshot,
) map[string]decimal.Decimal{

	"performance": PerformanceMetric,

	"weighted_performance": WeightedPerformanceMetric,

	"holdings": HoldingsMetric,
}

var Metrics = util.MapKeys(metricsMap)
