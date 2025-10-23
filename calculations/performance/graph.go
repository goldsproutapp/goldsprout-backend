package performance

import (
	"time"

	"github.com/goldsproutapp/goldsprout-backend/lib/processing"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"github.com/shopspring/decimal"
)

func GeneratePerformanceGraphInfo(snapshots []models.StockSnapshot) PerformanceGraphInfo {

	snapshotMapMerged, yearStartMap := processing.CreateMergedSnapshotMap(snapshots)

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
	return PerformanceGraphInfo{
		Performance: perfOut,
		Value:       valueOut,
		Cost:        costOut,
		YearToDate:  ytd,
	}
}
