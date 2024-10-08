package split

import (
	"fmt"
	"slices"
	"time"

	"github.com/goldsproutapp/goldsprout-backend/calculations/performance"
	"github.com/goldsproutapp/goldsprout-backend/constants"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"github.com/shopspring/decimal"
)

type splitMap map[string][]models.StockSnapshot

func IsSplitQueryValid(q models.SplitRequestQuery) bool {
	return slices.Contains(performance.Targets, q.Compare) &&
		slices.Contains(performance.Targets, q.Across)
}

func CategoriseSnapshots(snapshots []models.StockSnapshot, categoryKey string,
) splitMap {
	grouped := splitMap{}
	for _, snapshot := range snapshots {
		category := performance.GetKeyFromSnapshot(snapshot, categoryKey)
		if util.ContainsKey(grouped, category) {
			grouped[category] = append(grouped[category], snapshot)
		} else {
			grouped[category] = []models.StockSnapshot{snapshot}
		}
	}
	return grouped
}

func CalculateSplit(grouped splitMap) map[string]decimal.Decimal {
	totals := map[string]decimal.Decimal{}
	total := decimal.NewFromInt(0)
	for category, snapshots := range grouped {
		timeMap := map[string]time.Time{}
		valueMap := map[string]decimal.Decimal{}
		for _, snapshot := range snapshots {
			key := fmt.Sprintf("%d:%d", snapshot.AccountID, snapshot.StockID)
			latest, existsLatest := timeMap[key]
			if !existsLatest || snapshot.Date.Compare(latest) == 1 {
				timeMap[key] = snapshot.Date
				valueMap[key] = snapshot.Value
			}
		}
		values := make([]decimal.Decimal, 0, len(valueMap))
		for _, v := range valueMap {
			values = append(values, v)
		}
		sum := decimal.Sum(values[0], values[1:]...)
		totals[category] = sum
		total = total.Add(sum)
	}
	split := map[string]decimal.Decimal{}
	for category, sum := range totals {
		split[category] = sum.
			Div(total).
			Mul(decimal.NewFromInt(100)).
			Truncate(constants.PERFORMANCE_DECIMAL_DIGITS)
	}
	return split
}
