package extraction

import (
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"github.com/shopspring/decimal"
)

var singlePropGetters = map[string]func(models.StockSnapshot) string{
	"person": func(snapshot models.StockSnapshot) string {
		/*
			 TODO: ideally categorisation will be by id and display by name,
						but currently the same value is required for both.
		*/
		return snapshot.User.FirstName + " " + snapshot.User.LastName
	},
	"provider": func(snapshot models.StockSnapshot) string {
		return snapshot.Stock.Provider.Name
	},
	"account": func(snapshot models.StockSnapshot) string {
		return snapshot.Account.Name
	},
	"sector": func(snapshot models.StockSnapshot) string {
		return snapshot.Stock.Sector
	},
	"region": func(snapshot models.StockSnapshot) string {
		return snapshot.Stock.Region
	},
	"stock": func(snapshot models.StockSnapshot) string {
		return snapshot.Stock.Name
	},
	"all": func(_ models.StockSnapshot) string {
		return ""
	},
}

func SnapshotPropertyExtractionFunction(property string) func(models.StockSnapshot) string {
	return singlePropGetters[property]
}

func ExtractPropertyFromSnapshot(property string, snapshot models.StockSnapshot) string {
	return SnapshotPropertyExtractionFunction(property)(snapshot)
}

var multiPropGetters = map[string]func(models.StockSnapshot) []string{
	"class": func(snapshot models.StockSnapshot) []string {
		return util.MapKeys(snapshot.Stock.ClassCompositionMap)
	},
}

func SnapshotCompositePropertyExtractionFunction(property string) func(models.StockSnapshot) []string {
	return multiPropGetters[property]
}

func ExtractCompositePropertyFromSnapshot(property string, snapshot models.StockSnapshot) []string {
	return SnapshotCompositePropertyExtractionFunction(property)(snapshot)
}

func SplitSnapshotValueByPercentage(snapshot models.StockSnapshot, percentage decimal.Decimal) models.StockSnapshot {
	pct := percentage.Div(decimal.NewFromInt(100))
	return models.StockSnapshot{
		ID:                     snapshot.ID,
		User:                   snapshot.User,
		UserID:                 snapshot.UserID,
		Account:                snapshot.Account,
		AccountID:              snapshot.AccountID,
		Date:                   snapshot.Date,
		Stock:                  snapshot.Stock,
		StockID:                snapshot.StockID,
		Units:                  snapshot.Units.Mul(pct),
		Price:                  snapshot.Price,
		Cost:                   snapshot.Cost.Mul(pct),
		Value:                  snapshot.Value.Mul(pct),
		ChangeToDate:           snapshot.ChangeToDate.Mul(pct),
		ChangeSinceLast:        snapshot.ChangeSinceLast.Mul(pct),
		NormalisedPerformance:  snapshot.NormalisedPerformance,
		TransactionAttribution: snapshot.TransactionAttribution,
	}
}

var compositionSplitters = map[string]func(models.StockSnapshot, string) models.StockSnapshot{
	"class": func(snapshot models.StockSnapshot, key string) models.StockSnapshot {
		s := SplitSnapshotValueByPercentage(snapshot, snapshot.Stock.ClassCompositionMap[key])
		s.Stock.ClassCompositionMap = map[string]decimal.Decimal{key: decimal.NewFromInt(100)}
		return s
	},
}

func GetKeysFromSnapshot(snapshot models.StockSnapshot, key string) []string {
	extractionFunction := SnapshotPropertyExtractionFunction(key)
	if extractionFunction != nil {
		return util.Only(extractionFunction(snapshot))
	} else {
		return ExtractCompositePropertyFromSnapshot(key, snapshot)
	}
}

func GetContributionForCategory(snapshot models.StockSnapshot, key string, category string) models.StockSnapshot {
	if len(GetKeysFromSnapshot(snapshot, key)) == 1 {
		return snapshot
	}
	return compositionSplitters[key](snapshot, category)
}

func SingleTargets() []string {
	return util.MapKeys(singlePropGetters)
}

func MultiTargets() []string {
	return util.MapKeys(multiPropGetters)
}

func AllTargets() []string {
	return append(SingleTargets(), MultiTargets()...)
}
