package auth

import "github.com/patrickjonesuk/investment-tracker/models"

func GetAllowedUsers(user models.User, requireRead bool, requireWrite bool) []uint {
	var userIds []uint
	for _, perm := range user.AccessPermissions {
		uid := perm.AccessFor.ID
		if (perm.Read || !requireRead) && (perm.Write || !requireWrite) {
			userIds = append(userIds, uid)
		}
	}
	return append(userIds, user.ID) // user always has permissions for themselves.
}

func HasAccessPerm(user models.User, forUser uint, requireRead bool, requireWrite bool) bool {
    if user.IsAdmin {
        return true
    }
	for _, perm := range user.AccessPermissions {
		uid := perm.AccessFor.ID
		if uid != forUser {
			continue
		}
		return (perm.Read || !requireRead) && (perm.Write || !requireWrite)
	}
	return false
}
