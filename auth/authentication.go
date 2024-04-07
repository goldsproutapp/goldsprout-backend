package auth

import (
	"errors"

	"github.com/patrickjonesuk/investment-tracker/constants"
	"github.com/patrickjonesuk/investment-tracker/models"
	"gorm.io/gorm"
)

func AuthenticateToken(db *gorm.DB, token string, preload ...string) (models.User, error) {
	tokenHash := Hash(token)
	var user models.User = models.User{}
	qry := db.Where(models.User{TokenHash: tokenHash})
	for _, join := range preload {
		qry.Preload(join)
	}
	result := qry.First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return models.User{}, result.Error
	}
	return user, nil
}

func AuthenticateUnamePw(db *gorm.DB, uname string, password string) (models.User, error) {
	var user models.User = models.User{}
	result := db.Where(models.User{Email: uname}).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) || !ValidatePassword(password, user.PasswordHash) {
		return models.User{}, errors.New("Invalid username or password")
	}
	return user, nil
}

func GenerateToken() string {
    return GenerateUID(constants.TOKEN_LENGTH)
}

func CreateToken(db *gorm.DB, user models.User) string {
    token := GenerateToken()
    user.TokenHash = Hash(token)
    db.Save(&user)
    return token
}
