package reports

import (
	"time"

	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/shopspring/decimal"
)

type ReportQuery struct {
}

type ReportTransaction struct {
	Date        time.Time       `json:"date"`
	StockID     uint            `json:"stock_id"`
	AccountID   uint            `json:"account_id"`
	Value       decimal.Decimal `json:"value"`
	Units       decimal.Decimal `json:"units"`
	Price       decimal.Decimal `json:"price"`
	ValueAfter  decimal.Decimal `json:"value_after"`
	Attribution uint            `json:"attribution"`
}

type Report struct {
	StartValue  decimal.Decimal `json:"start_value"`
	EndValue    decimal.Decimal `json:"end_value"`
	GrossChange decimal.Decimal `json:"gross_change"`

	PurchaseTotal decimal.Decimal `json:"purchase_total"`
	SellTotal     decimal.Decimal `json:"sell_total"`
	NetCashflow   decimal.Decimal `json:"net_cashflow"`

	TotalGain decimal.Decimal `json:"total_gain"`

	Transactions []ReportTransaction `json:"transactions"`

	ExpectedFees decimal.Decimal `json:"expected_fees"`
	TotalFeePaid decimal.Decimal `json:"total_fee_paid"`

	TotalIncome   decimal.Decimal `json:"total_income"`
	SnapshotCount int             `json:"snapshot_count"`
}

type AggregatedSnapshotsMap struct {
	Snapshots       map[string][]models.StockSnapshot      // StockSnapshot.key() -> []StockSnapshot
	AccountPrevious map[uint]map[uint]models.StockSnapshot // AccountID -> StockID -> []StockSnapshot (penultimate snapshot list for account)
	AccountLast     map[uint]time.Time                     // AccountID -> Date (latest snapshot date for account)
}
