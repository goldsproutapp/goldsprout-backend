package calculations

import (
	"time"

	"github.com/patrickjonesuk/investment-tracker-backend/constants"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/shopspring/decimal"
)

func CalculateValueChange(value decimal.Decimal, prevSnapshot *models.StockSnapshot) decimal.Decimal {
	if prevSnapshot == nil {
		return decimal.NewFromInt(0)
	}
	return value.Sub(prevSnapshot.Value)
}

func CalculateNormalisedPerformance(
	price decimal.Decimal,
	prevSnapshot *models.StockSnapshot,
	date time.Time,
) decimal.Decimal {
	if prevSnapshot == nil {
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
