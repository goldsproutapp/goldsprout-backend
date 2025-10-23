package reports

import (
	"slices"

	"github.com/goldsproutapp/goldsprout-backend/lib/extraction"
	"github.com/goldsproutapp/goldsprout-backend/lib/extraction/times"
	"github.com/goldsproutapp/goldsprout-backend/models"
)

func IsReportQueryValid(query models.ReportRequestQuery) bool {
	return slices.Contains(extraction.TimeKeys(times.ReportExtractionSet()), query.Period)
}
