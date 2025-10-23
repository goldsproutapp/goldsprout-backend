package times

import (
	"strconv"
	"time"

	"github.com/goldsproutapp/goldsprout-backend/util"
)

func FocusYear(year string) []string {
	yearNum := util.ParseIntOrDefault(year, -1)
	start := time.Date(yearNum, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(yearNum, time.December, 31, 23, 0, 0, 0, time.UTC)
	return []string{"months", strconv.FormatInt(start.Unix(), 10), strconv.FormatInt(end.Unix(), 10)}
}

func FocusMonth(month string) []string {
	return []string{}
}
