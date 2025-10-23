package metrics

import (
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"github.com/shopspring/decimal"
)

type PerformanceMetricFunction func(
	timeMap map[string][]models.StockSnapshot,
) map[string]decimal.Decimal

type PerformanceMetricMeta struct {
	PermitLimited bool
	SummaryLabel  string
}

var metricsMap = map[string]PerformanceMetricFunction{

	"performance": PerformanceMetric,

	"weighted_performance": WeightedPerformanceMetric,

	"growth": GrowthMetric,

	"holdings": HoldingsMetric,

	"gains": GainsMetric,
}

func MetricFunctionByName(name string) PerformanceMetricFunction {
	return metricsMap[name]
}

func GetMetricNames() []string {
	return util.MapKeys(metricsMap)
}

var metricMeta = map[string]PerformanceMetricMeta{
	"performance": PerformanceMetricMeta{
		PermitLimited: true,
		SummaryLabel:  "Average",
	},
	"weighted_performance": PerformanceMetricMeta{
		PermitLimited: true,
		SummaryLabel:  "Average",
	},
	"growth": PerformanceMetricMeta{
		PermitLimited: true,
		SummaryLabel:  "Total",
	},
	"holdings": PerformanceMetricMeta{
		PermitLimited: false,
		SummaryLabel:  "Latest",
	},
	"gains": PerformanceMetricMeta{
		PermitLimited: false,
		SummaryLabel:  "Total",
	},
}

func GetMetricMetaByName(name string) PerformanceMetricMeta {
	return metricMeta[name]
}
