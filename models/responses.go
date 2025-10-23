package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type OverviewResponseUserEntry struct {
	TotalValue    decimal.Decimal `json:"total_value,omitempty"`
	AllTimeChange decimal.Decimal `json:"all_time_change,omitempty"`
	NumProviders  int             `json:"num_providers,omitempty"`
	NumStocks     int             `json:"num_stocks,omitempty"`
	LastSnapshot  time.Time       `json:"last_snapshot,omitempty"`
}

type OverviewResponse struct {
	OverviewResponseUserEntry
	Users map[string]OverviewResponseUserEntry `json:"users,omitempty"`
	AUM   decimal.Decimal                      `json:"aum,omitempty"`
}

type AccountReponse struct {
	Account
	Value      decimal.Decimal `json:"value,omitempty"`
	StockCount uint            `json:"stock_count,omitempty"`
}

type HoldingInfo struct {
	Value decimal.Decimal `json:"value,omitempty"`
	Units decimal.Decimal `json:"units,omitempty"`
}

func (i HoldingInfo) Merge(other HoldingInfo) HoldingInfo {
	return HoldingInfo{
		Value: i.Value.Add(other.Value),
		Units: i.Units.Add(other.Units),
	}
}
