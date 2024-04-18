package database

import (
	"slices"

	"github.com/patrickjonesuk/investment-tracker-backend/auth"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/util"
	"gorm.io/gorm"
)

func FetchPerformanceData(db *gorm.DB, user models.User, filter models.PerformanceFilter) []models.StockSnapshot {
	uids := auth.GetAllowedUsers(user, true, false)
	if user.IsAdmin {
		uids = util.UserIDs(GetAllUsers(db))
	}
	if len(filter.Users) > 0 {
		intersection := make([]uint, 0)
		for _, uid := range uids {
			if slices.Contains(filter.Users, uid) {
				intersection = append(intersection, uid)
			}
		}
		uids = intersection
	}
	qry := db.Model(&models.StockSnapshot{}).
		Order("date").
		Joins("User").
        Joins("Stock").
        Joins("INNER JOIN providers ON provider_id = providers.id").
        Preload("Stock.Provider").
		Where("user_id IN ?", uids)
	if len(filter.Providers) > 0 {
		qry = qry.Where("Stock.provider_id IN ?", filter.Providers)
	}
	if len(filter.Regions) > 0 {
		qry = qry.Where("Stock.region IN ?", filter.Regions)
	}

	var snapshots []models.StockSnapshot
	qry.Find(&snapshots)
	return snapshots
}
