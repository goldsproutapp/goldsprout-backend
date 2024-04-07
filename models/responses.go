package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type OverviewResponseUserEntry struct {
	TotalValue   decimal.Decimal `json:"total_value,omitempty"`
	NumProviders int             `json:"num_providers,omitempty"`
	NumStocks    int             `json:"num_stocks,omitempty"`
	LastSnapshot time.Time       `json:"last_snapshot,omitempty"`
}

type OverviewResponse struct {
	OverviewResponseUserEntry
	Users map[string]OverviewResponseUserEntry `json:"users,omitempty"`
}

