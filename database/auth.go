package database

import (
	"slices"

	"github.com/goldsproutapp/goldsprout-backend/auth"
	"github.com/goldsproutapp/goldsprout-backend/models"
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
	if user.Trusted && slices.Contains(uids, user.ID) {
		return true
	}
	for _, uid := range uids {
		if uid != user.ID && !auth.HasAccessPerm(user, uid, false, true) {
			return false
		}
	}
	return true
}
