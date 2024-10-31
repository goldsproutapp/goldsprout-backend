package models

import "fmt"

func (s *StockSnapshot) Key() string {
	return fmt.Sprintf("%s:%s", s.AccountID, s.StockID)
}

func (i *PerformanceQueryInfo) GenerateSummary() bool {
	return i.TargetKey != "all"
}
