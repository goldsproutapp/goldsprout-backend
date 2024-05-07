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
	ID            uint   `json:"id"`
	UserID        uint   `json:"user_id"`
	Stock         Stock  `json:"stock"`
	StockID       uint   `json:"stock_id"`
	CurrentlyHeld bool   `json:"currently_held"`
	Notes         string `json:"notes"`
}

type StockSnapshot struct {
	ID                    uint            `json:"id"`
	User                  User            `json:"-"`
	UserID                uint            `json:"user_id"`
	Date                  time.Time       `json:"date"`
	Stock                 Stock           `json:"-"`
	StockID               uint            `json:"stock_id"`
	Units                 decimal.Decimal `json:"units"`
	Price                 decimal.Decimal `json:"price"`
	Cost                  decimal.Decimal `json:"cost"`
	Value                 decimal.Decimal `json:"value"`
	ChangeToDate          decimal.Decimal `json:"changeToDate"`
	ChangeSinceLast       decimal.Decimal `json:"changeSinceLast"`       // absolute change in value
	NormalisedPerformance decimal.Decimal `json:"normalisedPerformance"` // relative change in price per unit (normalised for 30 days)
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
