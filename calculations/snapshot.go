package calculations

import (
	"time"

	"github.com/goldsproutapp/goldsprout-backend/constants"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/shopspring/decimal"
)

func CalculateValueChange(totalChange decimal.Decimal, prevSnapshot *models.StockSnapshot) decimal.Decimal {
	if prevSnapshot == nil {
		return totalChange
	}
    return totalChange.Sub(prevSnapshot.ChangeToDate).Truncate(2)
}

func CalculateNormalisedPerformance(
	price decimal.Decimal,
	prevSnapshot *models.StockSnapshot,
	date time.Time,
) decimal.Decimal {
	if prevSnapshot == nil || prevSnapshot.Price.Equal(decimal.NewFromInt(0)) {
		return decimal.NewFromInt(0)
	}
	perfChange := price.Sub(prevSnapshot.Price).Div(prevSnapshot.Price)
	timeDelta := date.Sub(prevSnapshot.Date)
	normalised := perfChange.Div(
		decimal.NewFromFloat(timeDelta.Hours()),
	).Mul(
		decimal.NewFromInt(constants.PERFORMANCE_NORMALISATION_DAYS * 24))
	return normalised.Mul(decimal.NewFromInt(100)). // more useful as a percentage
							Truncate(constants.PERFORMANCE_DECIMAL_DIGITS)

}
