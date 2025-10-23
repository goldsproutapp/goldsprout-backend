package times

import "github.com/goldsproutapp/goldsprout-backend/models"

type TimeExtractionSet = map[string]func(models.StockSnapshot) string

func PerformanceTimeExtractionSet() TimeExtractionSet {
	return TimeExtractionSet{
		"years":  YearFormatter(),
		"months": MonthNameFormatter(),
	}
}

func ReportExtractionSet() TimeExtractionSet {
	return TimeExtractionSet{
		"annual":  YearFormatter(),
		"monthly": MonthYearFormatter(),
	}
}
