package database

import (
	"github.com/goldsproutapp/goldsprout-backend/auth"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"gorm.io/gorm"
)

func GetSnapshot(db *gorm.DB, id uint) (models.StockSnapshot, error) {
	var obj models.StockSnapshot
	result := db.Model(models.StockSnapshot{}).Where("id = ?", id).First(&obj)
	return obj, result.Error
}

func GetLatestSnapshots(userStocks []models.UserStock, db *gorm.DB) []*models.StockSnapshot {

	// PERF: Is there a way to combine this into a single query?
	snapshots := make([]*models.StockSnapshot, len(userStocks))
	for i, userStock := range userStocks {
		var snapshot models.StockSnapshot

		result := db.Model(models.StockSnapshot{}).Joins("Stock").
			Where("user_id = ? AND stock_id = ? AND account_id = ?",
				userStock.UserID, userStock.StockID, userStock.AccountID,
			).
			Order("date DESC").
			First(&snapshot)
		if !Exists(result) {
			snapshots[i] = nil
		} else {
			snapshots[i] = &snapshot
		}
	}

	return snapshots
}

func GetAccountSnapshots(accountID uint, db *gorm.DB, preload ...string) []models.StockSnapshot {
	var snapshots []models.StockSnapshot
	qry := db.Order("date").Where("account_id = ?", accountID)
	for _, join := range preload {
		qry = qry.Preload(join)
	}
	qry.Find(&snapshots)
	return snapshots
}

func GetAllVisibleSnapshots(user models.User, db *gorm.DB, permitLimited bool, preload ...string) []models.StockSnapshot {
	var snapshots []models.StockSnapshot
	qry := db.Order("date")
	if !user.IsAdmin {
		allowed_uids := auth.GetAllowedUsers(user, true, false, permitLimited)
		qry = qry.Where("user_id IN ?", allowed_uids)
	}
	for _, join := range preload {
		qry = qry.Preload(join)
	}
	qry.Find(&snapshots)
	return snapshots
}

func GetSnapshots(users []uint, stocks []uint, db *gorm.DB, preload ...string) []models.StockSnapshot {
	var snapshots []models.StockSnapshot
	qry := db.Order("date")
	if len(users) > 0 {
		qry = qry.Where("user_id IN ?", users)
	}
	if len(stocks) > 0 {
		qry = qry.Where("stock_id IN ?", stocks)
	}
	for _, join := range preload {
		qry = qry.Preload(join)
	}
	qry.Find(&snapshots)
	return snapshots
}

