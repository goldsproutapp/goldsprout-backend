package trends

import (
	"github.com/goldsproutapp/goldsprout-backend/calculations/trends/metrics"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/shopspring/decimal"
)

type PerformanceQueryInfo struct {
	TargetKey      string
	AgainstKey     string
	TimeKey        string
	MetricKey      string
	Meta           metrics.PerformanceMetricMeta
	MetricFunction metrics.PerformanceMetricFunction
	LatestOnly     bool
}

func (i *PerformanceQueryInfo) GenerateSummary() bool {
	return i.TargetKey != "all"
}

type CategoryPerformance struct {
	Totals map[string]decimal.Decimal            `json:"totals,omitempty"`
	Items  map[string]map[string]decimal.Decimal `json:"items,omitempty"`
}

type PerformanceResponse struct {
	TimePeriods []string                       `json:"time_periods,omitempty"`
	TimeFocus   [][]string                     `json:"time_focus,omitempty"`
	Data        map[string]CategoryPerformance `json:"data,omitempty"`
	SummaryRow  string                         `json:"summary_row"`
}

type PerformanceMap = map[string]map[string]map[string][]models.StockSnapshot

//                        ^col1      ^col2      ^time
