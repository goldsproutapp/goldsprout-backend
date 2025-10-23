package split

import (
	"slices"

	"github.com/goldsproutapp/goldsprout-backend/lib/extraction"
	"github.com/goldsproutapp/goldsprout-backend/models"
)

func IsSplitQueryValid(q models.SplitRequestQuery) bool {
	return slices.Contains(extraction.AllTargets(), q.Compare) &&
		slices.Contains(extraction.AllTargets(), q.Across)
}
