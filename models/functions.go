package models

import "fmt"

func (s *StockSnapshot) Key() string {
	return fmt.Sprintf("%s:%s", s.AccountID, s.StockID)
}
