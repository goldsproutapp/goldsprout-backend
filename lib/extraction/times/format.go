package times

import "github.com/goldsproutapp/goldsprout-backend/models"

func DateFormatter(fmt string) func(models.StockSnapshot) string {
	return func(snapshot models.StockSnapshot) string {
		return snapshot.Date.Format(fmt)
	}
}

func YearFormatter() func(models.StockSnapshot) string {
	return DateFormatter("2006")
}

func MonthNameFormatter() func(models.StockSnapshot) string {
	return DateFormatter("January")
}

func MonthYearFormatter() func(models.StockSnapshot) string {
	return DateFormatter("")
}
