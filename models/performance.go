package models

import (
	"github.com/shopspring/decimal"
)

type PerformanceQueryInfo struct {
	TargetKey  string
	AgainstKey string
	TimeKey    string
	MetricKey  string
}

type CategoryPerformance struct {
	Totals map[string]decimal.Decimal            `json:"totals,omitempty"`
	Items  map[string]map[string]decimal.Decimal `json:"items,omitempty"`
}

type PerformanceResponse struct {
	TimePeriods []string                       `json:"time_periods,omitempty"`
	Data        map[string]CategoryPerformance `json:"data,omitempty"`
}

type PerformanceMap = map[string]map[string]map[string][]StockSnapshot
//                        ^col1      ^col2      ^time
