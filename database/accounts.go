package database

import (
	"github.com/goldsproutapp/goldsprout-backend/auth"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"gorm.io/gorm"
)

func GetVisibleAccounts(db *gorm.DB, user models.User, permitLimited bool) ([]models.Account, error) {
	var accounts []models.Account
	qry := db.Model(&models.Account{})
	if !user.IsAdmin {
		uids := auth.GetAllowedUsers(user, true, false, permitLimited)
		qry = qry.Where("user_id IN ?", uids)
	}
	res := qry.Find(&accounts)
	return accounts, res.Error
}

func GetAccount(db *gorm.DB, id uint) (models.Account, error) {
	var account models.Account
	res := db.Model(&models.Account{}).Where("id = ?", id).First(&account)
	return account, res.Error
}

func GetStocksForAccount(db *gorm.DB, accountID uint) ([]models.UserStock, error) {
	var stocks []models.UserStock
	res := db.Model(&models.UserStock{}).
		Where("account_id = ? AND currently_held = true", accountID).
		Find(&stocks)
	return stocks, res.Error
}
