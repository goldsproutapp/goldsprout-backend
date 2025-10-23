package database

import (
	"slices"

	"github.com/goldsproutapp/goldsprout-backend/auth"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"gorm.io/gorm"
)

func GetFilteredSnapshots(db *gorm.DB, user models.User, filter StockFilter, permitLimited bool) []models.StockSnapshot {
	uids := auth.GetAllowedUsers(user, true, false, permitLimited)
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
		Joins("Account").
		Preload("Stock.Provider").
		Where("stock_snapshots.user_id IN ?", uids)
	if len(filter.Providers) > 0 {
		qry = qry.Where("Stock.provider_id IN ?", filter.Providers)
	}
	if len(filter.Regions) > 0 {
		qry = qry.Where("Stock.region IN ?", filter.Regions)
	}
	if len(filter.Accounts) > 0 {
		qry = qry.Where("Account.name IN ?", filter.Accounts)
	}
	if filter.LowerDate.Unix() != 0 {
		qry = qry.Where("date > ?", filter.LowerDate)
	}
	if filter.UpperDate.Unix() != 0 {
		qry = qry.Where("date < ?", filter.UpperDate)
	}

	var snapshots []models.StockSnapshot
	qry.Find(&snapshots)
	return snapshots
}
