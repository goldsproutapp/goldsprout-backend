package auth

import (
	"github.com/goldsproutapp/goldsprout-backend/models"
)

func GetAllowedUsers(user models.User, requireRead bool, requireWrite bool) []uint {
	var userIds []uint
	for _, perm := range user.AccessPermissions {
		uid := perm.AccessForID
		if (perm.Read || !requireRead) && (perm.Write || !requireWrite) {
			userIds = append(userIds, uid)
		}
	}
	return append(userIds, user.ID) // user always has permissions for themselves.
}

func HasAccessPerm(user models.User, forUser uint, requireRead bool, requireWrite bool) bool {
	if user.IsAdmin || user.ID == forUser {
		return true
	}
	for _, perm := range user.AccessPermissions {
		if perm.AccessForID != forUser {
			continue
		}
		return (perm.Read || !requireRead) && (perm.Write || !requireWrite)
	}
	return false
}
