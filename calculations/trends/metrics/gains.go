package metrics

import (
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/shopspring/decimal"
)


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
	items[GetMetricMetaByName("gains").SummaryLabel] = total
	return items
}
