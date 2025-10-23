package reports

import "time"

var prevPeriod = map[string]func(time.Time) time.Time{
	"monthly": func(t time.Time) time.Time {
		return t.AddDate(0, 0, -t.Day())
	},
	"annual": func(t time.Time) time.Time {
		return time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	},
}

func GetPreviousTimePeriod(timeType string, period time.Time) time.Time {
	return prevPeriod[timeType](period)
}
