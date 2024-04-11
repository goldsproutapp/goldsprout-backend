package models

import "time"

type User struct {
	ID           uint   `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	PasswordHash string `json:"-"`
	TokenHash    string `json:"-"`

	Snapshots           []StockSnapshot      `gorm:"foreignKey:UserID" json:"-"`
	RegularTransactions []RegularTransaction `gorm:"foreignKey:UserID" json:"-"`
	SingleTransactions  []SingleTransaction  `gorm:"foreignKey:UserID" json:"-"`

	IsAdmin           bool               `json:"is_admin"`
	AccessPermissions []AccessPermission `gorm:"foreignKey:UserID" json:"access_permissions"`
	InvitationToken   string             `json:"-"`
	Active            bool               `json:"active"`

	ClientOpts string    `json:"client_options"` // Likely for colour scheme etc. but the client can do whatever with this.
	CreatedAt  time.Time `json:"created_at"`
}

func (u User) Name() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) ApplyUpdate(update UserUpdateInfo) {
	u.FirstName = update.FirstName
	u.LastName = update.LastName
}

func (u User) PublicInfo() PublicUserInfo {
	return PublicUserInfo{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		IsAdmin:   u.IsAdmin,
	}
}

type PublicUserInfo struct {
	ID        uint   `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
	IsAdmin   bool   `json:"is_admin,omitempty"`
}

type UserUpdateInfo struct {
	FirstName string `binding:"required" json:"first_name,omitempty"`
	LastName  string `binding:"required" json:"last_name,omitempty"`
}
