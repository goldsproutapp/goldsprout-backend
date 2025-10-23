package times

import (
	"sort"

	"github.com/goldsproutapp/goldsprout-backend/constants"
	"github.com/goldsproutapp/goldsprout-backend/util"
)

func ListYears(years []string) []string {
	sort.Slice(years, func(a, b int) bool {
		errList := []error{}
		return util.ParseUint(years[a], &errList) < util.ParseUint(years[b], &errList)
	})
	return years
}

func ListMonths(months []string) []string {
	return constants.MONTHS
}

