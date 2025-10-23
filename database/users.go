package database

import (
	"github.com/goldsproutapp/goldsprout-backend/models"
	"gorm.io/gorm"
)

func GetAllUsers(db *gorm.DB, preload ...string) []models.User {
	var users []models.User
	qry := db.Model(&models.User{})
	for _, join := range preload {
		qry = qry.Preload(join)
	}
	qry.Find(&users)
	return users
}


func GetDemoUser(db *gorm.DB) models.User {
	var user models.User
	db.Model(&models.User{}).Where("is_demo_user = TRUE").First(&user)
	return user
}
