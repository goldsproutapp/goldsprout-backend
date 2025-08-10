package models

import (
	"github.com/goldsproutapp/goldsprout-backend/constants"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func (s *Stock) BeforeSave(tx *gorm.DB) error {

	objs := []ClassCompositionEntry{}
	labels := []string{}
	for k, v := range s.ClassCompositionMap {
		objs = append(objs, ClassCompositionEntry{StockID: s.ID, Label: k, Percentage: v})
		labels = append(labels, k)
	}
	tx.Save(&objs)
	tx.Where("stock_id = ?", s.ID).Where("label NOT IN ?", labels).Delete(&ClassCompositionEntry{})
	return nil
}

func (s *Stock) AfterFind(tx *gorm.DB) error {
	s.ClassCompositionMap = map[string]decimal.Decimal{}
	if len(s.classCompositionObjects) == 0 { // we're wasting a query if it has been fetched but there are no entries.
		tx.Model(&ClassCompositionEntry{}).Where("stock_id = ?", s.ID).Find(&(s.classCompositionObjects))
	}
	if len(s.classCompositionObjects) == 0 {
		s.ClassCompositionMap[constants.DEFAULT_CLASS_NAME] = decimal.NewFromInt(100)
	} else {
		for _, obj := range s.classCompositionObjects {
			s.ClassCompositionMap[obj.Label] = obj.Percentage
		}
	}
	return nil
}

func (u *UserStock) AfterFind(tx *gorm.DB) error {
	if u.Stock.ID != 0 {
		u.Stock.AfterFind(tx)
	}
	return nil
}
