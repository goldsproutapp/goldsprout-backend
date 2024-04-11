package database

import (
	"errors"
	"strconv"
	"time"

	"github.com/patrickjonesuk/investment-tracker-backend/auth"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/util"
	"github.com/patrickjonesuk/investment-tracker-backend/util/tristate"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func GetUserStocks(authUser models.User, db *gorm.DB, heldBy []uint, currentlyHeld tristate.Tristate) []models.UserStock {
	var userStocks []models.UserStock
	query := db.Model(&models.UserStock{}).Joins("Stock")
	if len(heldBy) > 0 {
		query = query.Where("user_id IN ?", heldBy)
	}
	if !currentlyHeld.IsNone() {
		query.Where("currently_held = ?", currentlyHeld.GetBoolValue(false))
	}
	if !authUser.IsAdmin {
		query = query.Where("user_id IN ?", auth.GetAllowedUsers(authUser, true, false))
	}
	query.Find(&userStocks)
	return userStocks
}

func GetVisibleStockList(user models.User, db *gorm.DB) []models.UserStock {
	return GetUserStocks(user, db, []uint{}, tristate.None())
}

func GetHeldStocks(user models.User, db *gorm.DB) []models.UserStock {
	return GetUserStocks(user, db, []uint{}, tristate.True())
}

func GetAllSnapshots(user models.User, db *gorm.DB) []models.StockSnapshot {
	var snapshots []models.StockSnapshot
	allowed_uids := auth.GetAllowedUsers(user, true, false)
	db.
		Where("user_id IN ?", allowed_uids).
		Order("date").
		Find(&snapshots)
	return snapshots
}

func GetLatestSnapshots(userStocks []models.UserStock, db *gorm.DB) []*models.StockSnapshot {
	uid_stockids := util.Map(userStocks, func(stock models.UserStock) []uint {
		return []uint{stock.UserID, stock.StockID}
	})

	// PERF: Is there a way to combine this into a single query?
	snapshots := make([]*models.StockSnapshot, len(uid_stockids))
	for i, pair := range uid_stockids {
		uid, sid := pair[0], pair[1]
		var snapshot models.StockSnapshot

		result := db.Model(models.StockSnapshot{}).Joins("Stock").
			Where(map[string]interface{}{"user_id": uid, "stock_id": sid}). // cannot use struct query due to zero values
			Order("date DESC").
			First(&snapshot)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			snapshots[i] = nil
		} else {
			snapshots[i] = &snapshot
		}
	}

	return snapshots
}

func GetGlobalStockByName(db *gorm.DB, name string, provider uint) (models.Stock, error) {
	var obj models.Stock
	result := db.Model(models.Stock{}).Where(models.Stock{Name: name, ProviderID: provider}).First(&obj)
	return obj, result.Error
}

func GetUserStockByName(db *gorm.DB, uid uint, name string) (models.UserStock, error) {
	var obj models.UserStock
	result := db.Model(models.UserStock{}).Joins("INNER JOIN stocks on stocks.id = user_stocks.stock_id").Where("user_id = ? AND stocks.name = ?", uid, name).First(&obj)
	return obj, result.Error
}
func GetUsersHoldingStock(db *gorm.DB, stockID uint) ([]uint, error) {
	var userStocks []models.UserStock
	result := db.Model(models.UserStock{}).Where("stock_id = ?", stockID).Find(&userStocks)
	uids := make([]uint, len(userStocks))
	for i, userStock := range userStocks {
		uids[i] = userStock.UserID
	}
	return uids, result.Error
}

func GetProviders(db *gorm.DB) []models.Provider {
	var providers []models.Provider
	db.Find(&providers)
	return providers
}

func GetRegions(db *gorm.DB) []string {
	var regions []string
	db.Model(&models.Stock{}).Select("region").Distinct("region").Find(&regions)
	return regions
}

func GetSectors(db *gorm.DB) []string {
	var sectors []string
	db.Model(&models.Stock{}).Select("sector").Distinct("sector").Find(&sectors)
	return sectors
}

func GetOverview(db *gorm.DB, user models.User) models.OverviewResponse {
	uids := auth.GetAllowedUsers(user, true, false)
	if user.IsAdmin {
		uids = util.UserIDs(GetAllUsers(db))
	}
	overviews := map[string]models.OverviewResponseUserEntry{}
	for _, uid := range uids {
		if uid != user.ID {
			overviews[strconv.FormatInt(int64(uid), 10)] = GetOverviewForUser(db, uid)
		}
	}
	return models.OverviewResponse{
		OverviewResponseUserEntry: GetOverviewForUser(db, user.ID),
		Users:                     overviews,
	}
}

func GetOverviewForUser(db *gorm.DB, uid uint) models.OverviewResponseUserEntry {
	var userStocks []models.UserStock
	db.Model(&models.UserStock{}).Preload("stocks").Where("currently_held = true").Where("user_id = ?", uid).Find(&userStocks)
	snapshots := GetLatestSnapshots(userStocks, db)
	totalValue := decimal.NewFromInt(0)
	allTimeChange := decimal.NewFromInt(0)
	providers := util.NewHashSet[uint]()
	numStocks := len(userStocks)
	lastSnapshot := time.Unix(0, 0)
	for i, us := range userStocks {
		snapshot := snapshots[i]
		if snapshot == nil {
			continue
		}
		totalValue = totalValue.Add(snapshot.Value)
		allTimeChange = allTimeChange.Add(snapshot.ChangeToDate)
		providers.Add(us.Stock.ProviderID)
		if snapshot.Date.Compare(lastSnapshot) == 1 {
			lastSnapshot = snapshot.Date
		}
	}
	return models.OverviewResponseUserEntry{
		TotalValue:    totalValue,
		AllTimeChange: allTimeChange,
		NumStocks:     numStocks,
		NumProviders:  providers.Size(),
		LastSnapshot:  lastSnapshot,
	}
}

func GetAllUsers(db *gorm.DB, preload ...string) []models.User {
	var users []models.User
	qry := db.Model(&models.User{})
	for _, join := range preload {
		qry = qry.Preload(join)
	}
	db.Find(&users)
	return users
}
