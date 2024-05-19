package database

import (
	"github.com/patrickjonesuk/investment-tracker-backend/auth"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"gorm.io/gorm"
)

func CanModifyStock(db *gorm.DB, user models.User, id uint) bool {
	if user.IsAdmin {
		return true
	}
	/*
				If the user is not an admin, then they can only modify the stock
				If they have write permissions for every user who holds it
		        (or they are the only user who holds it)
	*/
	uids, err := GetUsersHoldingStock(db, id)
	if err != nil {
		return false
	}
	for _, uid := range uids {
		if uid != user.ID && !auth.HasAccessPerm(user, uid, false, true) {
			return false
		}
	}
	return true
}
