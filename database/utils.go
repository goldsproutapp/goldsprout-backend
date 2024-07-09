package database

import (
	"errors"

	"gorm.io/gorm"
)

func Exists(db *gorm.DB) bool {
	return !errors.Is(db.Error, gorm.ErrRecordNotFound)
}
