package database

import (
	"strconv"
	"time"

	"github.com/goldsproutapp/goldsprout-backend/auth"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)



func GetProviders(db *gorm.DB) []models.Provider {
	var providers []models.Provider
	db.Find(&providers)
	return providers
}

func GetRegions(db *gorm.DB) []string {
	var regions []string
	db.Model(&models.Stock{}).Select("region").Distinct("region").Find(&regions)
	return regions
}

func GetSectors(db *gorm.DB) []string {
	var sectors []string
	db.Model(&models.Stock{}).Select("sector").Distinct("sector").Find(&sectors)
	return sectors
}

func GetClasses(db *gorm.DB) []string {
	var classes []string
	db.Model(&models.ClassCompositionEntry{}).Select("label").Distinct("label").Find(&classes)
	return classes
}

func GetOverview(db *gorm.DB, user models.User) models.OverviewResponse {
	uids := auth.GetAllowedUsers(user, true, false, false)
	if user.IsAdmin {
		uids = util.UserIDs(GetAllUsers(db))
	}
	overviews := map[string]models.OverviewResponseUserEntry{}
	userOverview := GetOverviewForUser(db, user.ID)
	aum := userOverview.TotalValue
	for _, uid := range uids {
		if uid != user.ID {
			overview := GetOverviewForUser(db, uid)
			aum = aum.Add(overview.TotalValue)
			overviews[strconv.FormatInt(int64(uid), 10)] = overview
		}
	}
	return models.OverviewResponse{
		OverviewResponseUserEntry: userOverview,
		Users:                     overviews,
		AUM:                       aum,
	}
}

func GetOverviewForUser(db *gorm.DB, uid uint) models.OverviewResponseUserEntry {
	var userStocks []models.UserStock
	db.Model(&models.UserStock{}).Where("currently_held = true").Where("user_id = ?", uid).Preload("Stock").Find(&userStocks)
	snapshots := GetLatestSnapshots(userStocks, db)
	totalValue := decimal.NewFromInt(0)
	allTimeChange := decimal.NewFromInt(0)
	providers := util.NewHashSet[uint]()
	numStocks := len(userStocks)
	lastSnapshot := time.Unix(0, 0)
	for i, us := range userStocks {
		snapshot := snapshots[i]
		if snapshot == nil {
			continue
		}
		totalValue = totalValue.Add(snapshot.Value)
		allTimeChange = allTimeChange.Add(snapshot.ChangeToDate)
		providers.Add(us.Stock.ProviderID)
		if snapshot.Date.Compare(lastSnapshot) == 1 {
			lastSnapshot = snapshot.Date
		}
	}
	return models.OverviewResponseUserEntry{
		TotalValue:    totalValue,
		AllTimeChange: allTimeChange,
		NumStocks:     numStocks,
		NumProviders:  providers.Size(),
		LastSnapshot:  lastSnapshot,
	}
}
