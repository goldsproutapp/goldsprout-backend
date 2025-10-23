package metrics

import (
	"time"

	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"github.com/shopspring/decimal"
)

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
	items[GetMetricMetaByName("holdings").SummaryLabel] = items[latestTimePeriod]
	return items
}
