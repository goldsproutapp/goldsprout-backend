package database

import (
	"errors"
	"fmt"

	"github.com/goldsproutapp/goldsprout-backend/auth"
	"github.com/goldsproutapp/goldsprout-backend/config"
	"github.com/goldsproutapp/goldsprout-backend/email"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func DBConnString() string {
	user := config.RequiredEnv(config.ENVKEY_DBUSER)
	pw := config.RequiredEnv(config.ENVKEY_DBPW)
	host := config.EnvOrDefault(config.ENVKEY_DBHOST, DEFAULT_DB_HOST)
	port := config.EnvOrDefault(config.ENVKEY_DBPORT, DEFAULT_DB_PORT)
	name := config.EnvOrDefault(config.ENVKEY_DBNAME, DEFAULT_DB_NAME)
	connString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user, pw, host, port, name,
	)
	return connString
}

func CreateInitialAdminAccount(db *gorm.DB) {
	res := db.Where(&models.User{ID: 1}).First(&models.User{})
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		token := auth.GenerateToken()
		emailAddress := config.RequiredEnv(config.ENVKEY_ADMIN_EMAIL)
		user := models.User{
			ID:              1,
			Email:           emailAddress,
			PasswordHash:    "",
			InvitationToken: token,
			FirstName:       config.RequiredEnv(config.ENVKEY_ADMIN_FNAME),
			LastName:        config.RequiredEnv(config.ENVKEY_ADMIN_LNAME),
			IsAdmin:         true,
			Trusted:         true,
			Active:          false,
		}
		db.Create(&user)
		email.SendSetupInvitation(emailAddress, token)
	}
}

func CreateDemoAccount(db *gorm.DB) {
	if Exists(db.Where(&models.User{IsDemoUser: true}).First(&models.User{})) {
		return
	}
	user := models.User{
		Email:           config.EnvOrDefault(config.ENVKEY_DEMO_USER_EMAIL, "demo@example.com"),
		PasswordHash:    "",
		InvitationToken: "",
		FirstName:       config.EnvOrDefault(config.ENVKEY_DEMO_USER_FIRST_NAME, "Demo"),
		LastName:        config.EnvOrDefault(config.ENVKEY_DEMO_USER_LAST_NAME, "User"),
		IsAdmin:         true,
		Trusted:         true,
		Active:          true,
		IsDemoUser:      true,
	}
	db.Create(&user)
}

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(DBConnString()), &gorm.Config{})
	if err != nil {
		panic("failed to connect to db")
	}
	db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.Provider{},
		&models.Account{},
		&models.Stock{},
		&models.UserStock{},
		&models.StockSnapshot{},
		&models.RegularTransaction{},
		&models.SingleTransaction{},
		&models.AccessPermission{},
	)
	return db
}
