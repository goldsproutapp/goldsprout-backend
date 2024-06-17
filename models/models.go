package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Provider struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	CSVFormat string  `json:"csv_format"`
	AnnualFee float32 `json:"annual_fee,omitempty"`
}

type Account struct {
	ID         uint     `json:"id,omitempty"`
	Name       string   `json:"name,omitempty"`
	Provider   Provider `json:"provider,omitempty"`
	ProviderID uint     `json:"provider_id,omitempty"`
	User       User     `json:"user,omitempty"`
	UserID     uint     `json:"user_id,omitempty"`
}

type Stock struct {
	ID               uint     `json:"id,omitempty"`
	Name             string   `json:"name,omitempty"`
	Provider         Provider `json:"provider,omitempty"`
	ProviderID       uint     `json:"provider_id,omitempty"`
	Sector           string   `json:"sector,omitempty"`
	Region           string   `json:"region,omitempty"`
	StockCode        string   `json:"stock_code,omitempty"`
	NeedsAttention   bool     `json:"needs_attention,omitempty"`                              // If a stock is created automatically then it needs reviewing manually.
	TrackingStrategy string   `json:"tracking_strategy,omitempty" gorm:"default:DATA_IMPORT"` // DATA_IMPORT | VALUE_INPUT | API_DATA
	AnnualFee        float32  `json:"annual_fee,omitempty"`
}

type UserStock struct {
	ID            uint    `json:"id,omitempty"`
	UserID        uint    `json:"user_id,omitempty"`
	Stock         Stock   `json:"stock,omitempty"`
	StockID       uint    `json:"stock_id,omitempty"`
	Account       Account `json:"account,omitempty"`
	AccountID     uint    `json:"account_id,omitempty"`
	CurrentlyHeld bool    `json:"currently_held,omitempty"`
	Notes         string  `json:"notes,omitempty"`
}

type StockSnapshot struct {
	ID                    uint            `json:"id,omitempty"`
	User                  User            `json:"user,omitempty"`
	UserID                uint            `json:"user_id,omitempty"`
	Account               Account         `json:"account,omitempty"`
	AccountID             uint            `json:"account_id,omitempty"`
	Date                  time.Time       `json:"date,omitempty"`
	Stock                 Stock           `json:"stock,omitempty"`
	StockID               uint            `json:"stock_id,omitempty"`
	Units                 decimal.Decimal `json:"units,omitempty"`
	Price                 decimal.Decimal `json:"price,omitempty"`
	Cost                  decimal.Decimal `json:"cost,omitempty"`
	Value                 decimal.Decimal `json:"value,omitempty"`
	ChangeToDate          decimal.Decimal `json:"change_to_date,omitempty"`
	ChangeSinceLast       decimal.Decimal `json:"change_since_last,omitempty"`      // absolute change in value
	NormalisedPerformance decimal.Decimal `json:"normalised_performance,omitempty"` // relative change in price per unit (normalised for 30 days)
}

type RegularTransaction struct {
	ID       uint
	UserID   uint
	Stock    Stock
	StockID  uint
	Amount   decimal.Decimal
	First    time.Time
	Last     *time.Time // nullable
	Interval string     // daily | weekly | fortnightly | monthly | quarterly | yearly
}

type SingleTransaction struct {
	ID      uint
	UserID  uint
	Stock   Stock
	StockID uint
	Amount  decimal.Decimal
	Date    time.Time
}

type AccessPermission struct {
	ID          uint `json:"-"`
	UserID      uint `json:"user_id,omitempty"`
	AccessFor   User `json:"-"`
	AccessForID uint `json:"access_for_id,omitempty"`
	Read        bool `json:"read,omitempty"`
	Write       bool `json:"write,omitempty"`
}
