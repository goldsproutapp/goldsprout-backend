package extraction

import (
	"strconv"
	"time"

	"github.com/goldsproutapp/goldsprout-backend/lib/extraction/times"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
)

func TimeExtractionFunction(set times.TimeExtractionSet, timeType string) func(models.StockSnapshot) string {
	return set[timeType]
}
func ExtractTimeFromSnapshot(set times.TimeExtractionSet, timeType string, snapshot models.StockSnapshot) string {
	return TimeExtractionFunction(set, timeType)(snapshot)
}

func TimeKeys(set times.TimeExtractionSet) []string {
	return util.MapKeys(set)
}

var timeFilterValidators = map[string]func(time.Time, time.Time) []string{
	"years": func(start time.Time, end time.Time) []string {
		out := []string{}
		for i := start.Year(); i <= end.Year(); i++ {
			out = append(out, strconv.FormatInt(int64(i), 10))
		}
		return out
	},
}
var timeListGetters = map[string]func([]string) []string{
	"years":  times.ListYears,
	"months": times.ListMonths,
}

func GetTimeListFunction(timeType string) func([]string) []string {
	return timeListGetters[timeType]
}

func ExtractTimeList(timeType string, times []string) []string {
	return GetTimeListFunction(timeType)(times)
}

var timeFocus = map[string]func(string) []string{
	"years":  times.FocusYear,
	"months": times.FocusMonth,
}

func GetTimeFocusFunction(timeType string) func(string) []string {
	return timeFocus[timeType]
}

func GetTimeFocusKeys(timeType string, time string) []string {
	return GetTimeFocusFunction(timeType)(time)
}


