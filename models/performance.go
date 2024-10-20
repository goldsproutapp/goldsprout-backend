package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type PerformanceQueryInfo struct {
	TargetKey      string
	AgainstKey     string
	TimeKey        string
	MetricKey      string
	Meta           PerformanceMetricMeta
	MetricFunction PerformanceMetricFunction
}

type PerformanceMetricFunction func(
	timeMap map[string][]StockSnapshot,
) map[string]decimal.Decimal

type PerformanceMetricMeta struct {
	PermitLimited bool
	SummaryLabel  string
}

type CategoryPerformance struct {
	Totals map[string]decimal.Decimal            `json:"totals,omitempty"`
	Items  map[string]map[string]decimal.Decimal `json:"items,omitempty"`
}

type PerformanceResponse struct {
	TimePeriods []string                       `json:"time_periods,omitempty"`
	TimeFocus   [][]string                     `json:"time_focus"`
	Data        map[string]CategoryPerformance `json:"data,omitempty"`
}

type PerformanceMap = map[string]map[string]map[string][]StockSnapshot

//                        ^col1      ^col2      ^time

type StockFilter struct {
	Regions   []string
	Providers []uint
	Users     []uint
	Accounts  []string
	LowerDate time.Time
	UpperDate time.Time
}

type PerformanceGraphInfo struct {
	Value       map[time.Time]decimal.Decimal `json:"value,omitempty"`
	Performance map[time.Time]decimal.Decimal `json:"performance,omitempty"`
	YearToDate  decimal.Decimal               `json:"year_to_date,omitempty"`
}
