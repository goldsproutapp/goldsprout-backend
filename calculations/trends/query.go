package trends

import (
	"slices"

	"github.com/goldsproutapp/goldsprout-backend/calculations/trends/metrics"
	"github.com/goldsproutapp/goldsprout-backend/lib/extraction"
	"github.com/goldsproutapp/goldsprout-backend/lib/extraction/times"
)

func IsPerformanceQueryValid(p PerformanceQueryInfo) bool {
	// TODO: split keys (eg. class) can be valid for some metrics eg. holdings.
	return slices.Contains(extraction.SingleTargets(), p.TargetKey) &&
		slices.Contains(extraction.SingleTargets(), p.AgainstKey) &&
		slices.Contains(metrics.GetMetricNames(), p.MetricKey) &&
		slices.Contains(extraction.TimeKeys(times.PerformanceTimeExtractionSet()), p.TimeKey)
}

func SetQueryMeta(p *PerformanceQueryInfo) {
	p.Meta = metrics.GetMetricMetaByName(p.MetricKey)
	p.MetricFunction = metrics.MetricFunctionByName(p.MetricKey)
}

