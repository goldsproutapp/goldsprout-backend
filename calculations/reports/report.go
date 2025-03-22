package reports

import (
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/goldsproutapp/goldsprout-backend/constants"
	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AggregatedSnapshotsMap struct {
	Snapshots       map[string][]models.StockSnapshot      // StockSnapshot.key() -> []StockSnapshot
	AccountPrevious map[uint]map[uint]models.StockSnapshot // AccountID -> StockID -> []StockSnapshot (penultimate snapshot list for account)
	AccountLast     map[uint]time.Time                     // AccountID -> Date (latest snapshot date for account)
}

var Periods = []string{"monthly", "annual"}

var TimeGetters = map[string]func(models.StockSnapshot) string{
	"monthly": func(s models.StockSnapshot) string {
		return s.Date.Format("Jan 2006")
	},
	"annual": func(s models.StockSnapshot) string {
		return fmt.Sprintf("%v", s.Date.Year())
	},
}
var PrevPeriod = map[string]func(time.Time) time.Time{
	"monthly": func(t time.Time) time.Time {
		return t.AddDate(0, 0, -t.Day())
	},
	"annual": func(t time.Time) time.Time {
		return time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	},
}

func IsReportQueryValid(query models.ReportRequestQuery) bool {
	return slices.Contains(Periods, query.Period)
}

func SplitSnapshots(timePeriodKey string, snapshots []models.StockSnapshot) (map[string][]models.StockSnapshot, []string) {
	splitSnapshotMap := map[string][]models.StockSnapshot{}
	timePeriodMap := map[string]time.Time{}
	for _, snapshot := range snapshots {
		period := TimeGetters[timePeriodKey](snapshot)
		if !util.ContainsKey(splitSnapshotMap, period) {
			splitSnapshotMap[period] = []models.StockSnapshot{}
		}
		splitSnapshotMap[period] = append(splitSnapshotMap[period], snapshot)
		if !util.ContainsKey(timePeriodMap, period) || timePeriodMap[period].After(snapshot.Date) {
			timePeriodMap[period] = snapshot.Date
		}
	}
	keys := util.MapKeys(timePeriodMap)
	sort.Slice(keys, func(i, j int) bool {
		return timePeriodMap[keys[i]].Before(timePeriodMap[keys[j]])
	})
	splitSnapshotMap["Total"] = snapshots
	return splitSnapshotMap, append(keys, "Total")
}

func AggregateSnapshots(db *gorm.DB, start time.Time, snapshots []models.StockSnapshot) AggregatedSnapshotsMap {
	aggregated := map[string][]models.StockSnapshot{}
	accountPrev := map[uint]map[uint]models.StockSnapshot{}
	accountLast := map[uint]time.Time{}
	for _, snapshot := range snapshots {

		key := snapshot.Key()
		if !util.ContainsKey(aggregated, key) {
			aggregated[key] = []models.StockSnapshot{}
		}
		if !util.ContainsKey(accountLast, snapshot.AccountID) || accountLast[snapshot.AccountID].Before(snapshot.Date) {
			accountLast[snapshot.AccountID] = snapshot.Date
		}
		if !util.ContainsKey(accountPrev, snapshot.AccountID) {
			var dateSnapshot models.StockSnapshot
			if database.Exists(db.Model(&models.StockSnapshot{}).
				Select("date").
				Where("account_id = ?", snapshot.AccountID).
				Where("date < ?", start).
				Order("date DESC").
				Limit(1).
				First(&dateSnapshot)) {
				var prevSnapshots []models.StockSnapshot
				db.Model(&models.StockSnapshot{}).
					Where("account_id = ?", snapshot.AccountID).
					Where("date = ?", dateSnapshot.Date).
					Find(&prevSnapshots)
				prevSnapshotMap := map[uint]models.StockSnapshot{}
				for _, s := range prevSnapshots {
					prevSnapshotMap[s.StockID] = s
				}
				accountPrev[snapshot.AccountID] = prevSnapshotMap
			} else {
				accountPrev[snapshot.AccountID] = map[uint]models.StockSnapshot{}
			}
		}
		aggregated[key] = append(aggregated[key], snapshot)
	}
	return AggregatedSnapshotsMap{
		Snapshots:       aggregated,
		AccountPrevious: accountPrev,
		AccountLast:     accountLast,
	}
}

func updateReportForHolding(aggregated AggregatedSnapshotsMap, key string, report *models.Report) {
	snapshots := aggregated.Snapshots[key]
	if len(snapshots) == 0 {
		return
	}

	stock := snapshots[0].StockID
	account := snapshots[0].AccountID
	var prevSnapshot models.StockSnapshot
	if util.ContainsKey(aggregated.AccountPrevious[account], stock) {
		prevSnapshot = aggregated.AccountPrevious[account][stock]
	} else {
		prevSnapshot = models.StockSnapshot{Date: snapshots[0].Date}
	}
	snapshotsWithPrev := append([]models.StockSnapshot{prevSnapshot}, snapshots...)
	last := snapshots[len(snapshots)-1]
	if aggregated.AccountLast[account].After(last.Date) {
		// Similar to the boundary-sale detection, if a holding has clearly been sold
		// (ie. there is no latest snapshot) but a zero-entry was not automatically
		// inserted at creation, then (ephemerally) insert one now.
		if !last.Value.IsZero() {
			snapshotsWithPrev = append(snapshotsWithPrev, models.StockSnapshot{
				Price:                  snapshots[len(snapshots)-1].Price,
				TransactionAttribution: constants.TransAttrBuySell,
				Date:                   last.Date,
			})
		}
	} else {
		report.EndValue = report.EndValue.Add(snapshots[len(snapshots)-1].Value)
	}
	transactions := []models.ReportTransaction{}
	fee := decimal.NewFromInt(0)
	for i, s := range snapshotsWithPrev[1:] {
		prev := snapshotsWithPrev[i]
		unitChange := s.Units.Sub(prev.Units)
		transactionValue := unitChange.Mul(s.Price).Div(decimal.NewFromInt(100)).Truncate(2)
		report.TotalGain = report.TotalGain.Add(s.ChangeSinceLast)
		feeRate := s.Stock.AnnualFee + s.Stock.Provider.AnnualFee
		x := decimal.NewFromFloat(s.Date.Sub(prev.Date).Hours() / 24)
		paidFee := x.Div(decimal.NewFromInt(365)).Mul(decimal.NewFromFloat32(feeRate).Div(decimal.NewFromInt(100)))
		fee = fee.Add(s.Value.Mul(paidFee).Truncate(2))
		if transactionValue.IsZero() || transactionValue.Abs().LessThan(decimal.NewFromInt(1)) {
			continue
		}
		transactions = append(transactions, models.ReportTransaction{
			Date:        s.Date,
			StockID:     stock,
			AccountID:   account,
			Value:       transactionValue,
			Units:       unitChange,
			Price:       s.Price,
			ValueAfter:  s.Value,
			Attribution: s.TransactionAttribution,
		})
		if s.TransactionAttribution == constants.TransAttrBuySell {
			if transactionValue.IsPositive() {
				report.PurchaseTotal = report.PurchaseTotal.Add(transactionValue)
			} else {
				report.SellTotal = report.SellTotal.Add(transactionValue.Abs())
			}
		} else if s.TransactionAttribution == constants.TransAttrIncomeFee {
			if transactionValue.IsPositive() {
				report.TotalIncome = report.TotalIncome.Add(transactionValue)
			} else {
				report.TotalFeePaid = report.TotalFeePaid.Add(transactionValue.Neg())
			}
		}
	}
	report.ExpectedFees = report.ExpectedFees.Add(fee)
	report.Transactions = append(report.Transactions, transactions...)
}

func GenerateReport(aggregated AggregatedSnapshotsMap) models.Report {
	zero := decimal.NewFromInt(0)
	report := models.Report{
		StartValue:  zero,
		EndValue:    zero,
		GrossChange: zero,

		PurchaseTotal: zero,
		SellTotal:     zero,
		NetCashflow:   zero,

		TotalGain: zero,

		TotalFeePaid: zero,
		ExpectedFees: zero,
		TotalIncome:  zero,

		Transactions:  []models.ReportTransaction{},
		SnapshotCount: 0,
	}
	for _, s := range aggregated.AccountPrevious {
		for _, snapshot := range s {
			report.StartValue = report.StartValue.Add(snapshot.Value)
		}

	}
	keys := util.NewHashSet[string]()
	for key, snapshots := range aggregated.Snapshots {
		keys.Add(key)
		report.SnapshotCount += len(snapshots)
		updateReportForHolding(aggregated, key, &report)
	}
	// If a sale occurs across a time-period bounary, there should be a zero-entry
	// in the later period. However, such snapshots have not always been automatically inserted,
	// so this will pick up any holdings which disappear over a boundary and assume
	// a sale to have occurred.
	for _, s := range aggregated.AccountPrevious {
		for _, snapshot := range s {
			key := snapshot.Key()
			if !keys.Has(key) {
				aggregated.Snapshots[key] = []models.StockSnapshot{models.StockSnapshot{
					AccountID:              snapshot.AccountID,
					StockID:                snapshot.StockID,
					Price:                  snapshot.Price,
					TransactionAttribution: constants.TransAttrBuySell,
					Date:                   snapshot.Date,
				}}
				updateReportForHolding(aggregated, key, &report)
			}
		}
	}

	report.NetCashflow = report.PurchaseTotal.Sub(report.SellTotal)
	report.GrossChange = report.EndValue.Sub(report.StartValue)

	return report
}
