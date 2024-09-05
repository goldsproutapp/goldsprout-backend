package auth

import (
	"errors"

	"github.com/goldsproutapp/goldsprout-backend/constants"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"gorm.io/gorm"
)

func AuthenticateToken(db *gorm.DB, token string, preload ...string) (models.Session, error) {
	tokenHash := Hash(token)

	var session models.Session
	res := db.Where(models.Session{TokenHash: tokenHash}).First(&session)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return models.Session{}, res.Error
	}
	return session, nil
}

func UserForSession(db *gorm.DB, session models.Session, preload ...string) (models.User, error) {

	var user models.User = models.User{}
	qry := db.Where(models.User{ID: session.UserID})
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

func CreateToken(db *gorm.DB, user models.User, client string) string {
	token := GenerateToken()
	session := models.Session{
		UserID:    user.ID,
		TokenHash: Hash(token),
		Client:    client,
	}
	db.Save(&session)
	return token
}
