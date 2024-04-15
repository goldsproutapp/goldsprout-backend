package database

import (
	"github.com/patrickjonesuk/investment-tracker-backend/auth"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/util"
	"gorm.io/gorm"
)

func FetchPerformanceData(db *gorm.DB, user models.User) []models.StockSnapshot {
	uids := auth.GetAllowedUsers(user, true, false)
    if (user.IsAdmin) {
        uids = util.UserIDs(GetAllUsers(db))
    }
	qry := db.Model(&models.StockSnapshot{}).
		Where("user_id IN ?", uids).
        Order("date").
		Preload("User").
		Preload("Stock").
		Preload("Stock.Provider")
	var snapshots []models.StockSnapshot
	qry.Find(&snapshots)
	return snapshots
}
