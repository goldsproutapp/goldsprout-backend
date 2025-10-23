package performance

import (
	"time"

	"github.com/shopspring/decimal"
)

type PerformanceGraphInfo struct {
	Value       map[time.Time]decimal.Decimal `json:"value,omitempty"`
	Cost        map[time.Time]decimal.Decimal `json:"cost,omitempty"`
	Performance map[time.Time]decimal.Decimal `json:"performance,omitempty"`
	YearToDate  decimal.Decimal               `json:"year_to_date,omitempty"`
}
