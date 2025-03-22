package models

import "fmt"

func (s *StockSnapshot) Key() string {
	return fmt.Sprintf("%v:%v", s.AccountID, s.StockID)
}

func (i *PerformanceQueryInfo) GenerateSummary() bool {
	return i.TargetKey != "all"
}
