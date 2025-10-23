package database

import "time"

type StockFilter struct {
	Regions   []string
	Providers []uint
	Users     []uint
	Accounts  []string
	LowerDate time.Time
	UpperDate time.Time
}
