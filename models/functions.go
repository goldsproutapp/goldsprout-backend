package models

import "fmt"

func (s *StockSnapshot) Key() string {
	return fmt.Sprintf("%v:%v", s.AccountID, s.StockID)
}

