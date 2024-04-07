package calculations

import (
	"time"

	"github.com/patrickjonesuk/investment-tracker/constants"
	"github.com/patrickjonesuk/investment-tracker/models"
	"github.com/patrickjonesuk/investment-tracker/util"
	"github.com/shopspring/decimal"
)

var metricsMap = map[string]func(
	timeMap map[string][]models.StockSnapshot,
) map[string]decimal.Decimal{

	"performance": func(
		timeMap map[string][]models.StockSnapshot,
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
	},

	"weighted_performance": func(
		timeMap map[string][]models.StockSnapshot,
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
	},
	"holdings": func(
		timeMap map[string][]models.StockSnapshot,
	) map[string]decimal.Decimal {

		items := map[string]decimal.Decimal{}
		total := decimal.NewFromInt(0)
		for timePeriod, snapshots := range timeMap {
			dateMap := map[uint]time.Time{}
			valueMap := map[uint]decimal.Decimal{}
			for _, snapshot := range snapshots {
                latest, existsLatest := dateMap[snapshot.StockID]
                if !existsLatest || snapshot.Date.Compare(latest) == 1 {
                    dateMap[snapshot.StockID] = snapshot.Date
                    valueMap[snapshot.StockID] = snapshot.Value
                }
			}
            timeTotal := decimal.NewFromInt(0)
            for _, value := range valueMap {
                timeTotal = timeTotal.Add(value)
            }
            total = timeTotal
			items[timePeriod] = timeTotal
		}
		items["Total"] = total
		return items
	},
}

var Metrics = util.MapKeys(metricsMap)
