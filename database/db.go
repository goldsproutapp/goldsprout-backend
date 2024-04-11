package database

import (
	"errors"
	"fmt"

	"github.com/patrickjonesuk/investment-tracker-backend/auth"
	"github.com/patrickjonesuk/investment-tracker-backend/config"
	"github.com/patrickjonesuk/investment-tracker-backend/email"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
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
			TokenHash:       "", // NOTE: I'm assuming that this is safe due to hashes being fixed length
			IsAdmin:         true,
			Active:          false,
		}
		db.Create(&user)
		email.SendSetupInvitation(emailAddress, token)
	}
}

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(DBConnString()), &gorm.Config{})
	if err != nil {
		panic("failed to connect to db")
	}
	db.AutoMigrate(
		&models.User{},
		&models.Provider{},
		&models.Stock{},
		&models.UserStock{},
		&models.StockSnapshot{},
		&models.RegularTransaction{},
		&models.SingleTransaction{},
		&models.AccessPermission{},
	)
	return db
}
