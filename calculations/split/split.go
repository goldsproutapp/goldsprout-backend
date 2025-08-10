package split

import (
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
		categories := performance.GetKeysFromSnapshot(snapshot, categoryKey)
		for _, category := range categories {
			s := performance.GetContributionForCategory(snapshot, categoryKey, category)
			if util.ContainsKey(grouped, category) {
				grouped[category] = append(grouped[category], s)
			} else {
				grouped[category] = []models.StockSnapshot{s}
			}
		}
	}
	return grouped
}

func CalculateSplit(grouped splitMap) map[string]decimal.Decimal {
	totals := map[string]decimal.Decimal{}
	total := decimal.NewFromInt(0)
	accountMap := map[uint]time.Time{}
	for category, snapshots := range grouped {
		timeMap := map[string]time.Time{}
		keyToAccountMap := map[string]uint{}
		valueMap := map[string]decimal.Decimal{}
		for _, snapshot := range snapshots {
			key := snapshot.Key()
			accountLatest, existsAccountLatest := accountMap[snapshot.AccountID]
			if !existsAccountLatest || snapshot.Date.After(accountLatest) {
				accountMap[snapshot.AccountID] = snapshot.Date
			}
			latest, existsLatest := timeMap[key]
			if !existsLatest || snapshot.Date.Compare(latest) == 1 && !accountMap[snapshot.AccountID].After(snapshot.Date) {
				timeMap[key] = snapshot.Date
				keyToAccountMap[key] = snapshot.AccountID
				valueMap[key] = snapshot.Value
			}
		}
		values := []decimal.Decimal{}
		for k, v := range valueMap {
			if !timeMap[k].Before(accountMap[keyToAccountMap[k]]) {
				values = append(values, v)
			}
		}
		var sum decimal.Decimal
		if len(values) == 0 {
			sum = decimal.NewFromInt(0)
		} else if len(values) == 1 {
			sum = values[0]
		} else {
			sum = decimal.Sum(values[0], values[1:]...)
		}
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
