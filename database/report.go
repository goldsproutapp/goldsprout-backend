package database

import (
	"time"

	"github.com/goldsproutapp/goldsprout-backend/models"
	"gorm.io/gorm"
)

func AccountSnapshotBeforeDate(db *gorm.DB, startDate time.Time, accountID uint, dst *models.StockSnapshot) bool {
	return Exists(db.Model(&models.StockSnapshot{}).
		Select("date").
		Where("account_id = ?", accountID).
		Where("date < ?", startDate).
		Order("date DESC").
		Limit(1).
		First(&dst))
}

func GetAccountSnapshotsForDate(db *gorm.DB, accountId uint, date time.Time) []models.StockSnapshot {
	var prevSnapshots []models.StockSnapshot
	db.Model(&models.StockSnapshot{}).
		Where("account_id = ?", accountId).
		Where("date = ?", date).
		Find(&prevSnapshots)
	return prevSnapshots
}
