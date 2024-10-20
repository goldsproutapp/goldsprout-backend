package database

import (
	"errors"
	"strconv"
	"time"

	"github.com/goldsproutapp/goldsprout-backend/auth"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"github.com/goldsproutapp/goldsprout-backend/util/tristate"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func GetUserStocks(authUser models.User, db *gorm.DB, heldBy []uint, currentlyHeld tristate.Tristate, permitLimited bool) []models.UserStock {
	var userStocks []models.UserStock
	query := db.Model(&models.UserStock{}).Joins("Stock")
	if len(heldBy) > 0 {
		query = query.Where("user_id IN ?", heldBy)
	}
	if !currentlyHeld.IsNone() {
		query.Where("currently_held = ?", currentlyHeld.GetBoolValue(false))
	}
	if !authUser.IsAdmin {
		query = query.Where("user_id IN ?", auth.GetAllowedUsers(authUser, true, false, permitLimited))
	}
	query.Find(&userStocks)
	return userStocks
}

func GetVisibleStockList(user models.User, db *gorm.DB, permitLimited bool) []models.UserStock {
	return GetUserStocks(user, db, []uint{}, tristate.None(), permitLimited)
}

func GetHeldStocks(user models.User, db *gorm.DB, permitLimited bool) []models.UserStock {
	return GetUserStocks(user, db, []uint{}, tristate.True(), permitLimited)
}

func GetUserSnapshots(user models.User, db *gorm.DB, preload ...string) []models.StockSnapshot {
	var snapshots []models.StockSnapshot
	qry := db.Order("date").Where("user_id = ?", user.ID)
	for _, join := range preload {
		qry = qry.Preload(join)
	}
	qry.Find(&snapshots)
	return snapshots
}

func GetAccountSnapshots(accountID uint, db *gorm.DB, preload ...string) []models.StockSnapshot {
	var snapshots []models.StockSnapshot
	qry := db.Order("date").Where("account_id = ?", accountID)
	for _, join := range preload {
		qry = qry.Preload(join)
	}
	qry.Find(&snapshots)
	return snapshots
}

func GetAllSnapshots(user models.User, db *gorm.DB, permitLimited bool, preload ...string) []models.StockSnapshot {
	var snapshots []models.StockSnapshot
	qry := db.Order("date")
	if !user.IsAdmin {
		allowed_uids := auth.GetAllowedUsers(user, true, false, permitLimited)
		qry = qry.Where("user_id IN ?", allowed_uids)
	}
	for _, join := range preload {
		qry = qry.Preload(join)
	}
	qry.Find(&snapshots)
	return snapshots
}

func GetLatestSnapshots(userStocks []models.UserStock, db *gorm.DB) []*models.StockSnapshot {

	// PERF: Is there a way to combine this into a single query?
	snapshots := make([]*models.StockSnapshot, len(userStocks))
	for i, userStock := range userStocks {
		var snapshot models.StockSnapshot

		result := db.Model(models.StockSnapshot{}).Joins("Stock").
			Where("user_id = ? AND stock_id = ? AND account_id = ?",
				userStock.UserID, userStock.StockID, userStock.AccountID,
			).
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

func GetUserStockByName(db *gorm.DB, uid uint, name string, providerID uint) (models.UserStock, error) {
	var obj models.UserStock
	result := db.Model(models.UserStock{}).
		Joins("INNER JOIN stocks on stocks.id = user_stocks.stock_id").
		Where("user_id = ? AND stocks.name = ? AND stocks.provider_id = ?",
			uid, name, providerID).
		First(&obj)
	return obj, result.Error
}
func GetUserStockByNameOrCode(db *gorm.DB, uid uint, name string, code string, providerID uint) (models.UserStock, error) {
	if code == "" {
		return GetUserStockByName(db, uid, name, providerID)
	}
	var obj models.UserStock
	result := db.Model(models.UserStock{}).
		Joins("INNER JOIN stocks on stocks.id = user_stocks.stock_id").
		Where("user_id = ? AND (stocks.stock_code = ? OR stocks.name = ?) AND stocks.provider_id = ?",
			uid, code, name, providerID).
		First(&obj)
	return obj, result.Error
}
func GetGlobalStockByNameOrCode(db *gorm.DB, name string, code string, providerID uint) (models.Stock, error) {
	if code == "" {
		return GetGlobalStockByName(db, name, providerID)
	}
	var obj models.Stock
	result := db.Model(models.Stock{}).
		Where("provider_id = ? AND (stock_code = ? OR name = ?)",
			providerID, code, name).
		First(&obj)
	return obj, result.Error
}

func GetUserStock(db *gorm.DB, uid uint, stockID uint, accountID uint) (models.UserStock, error) {
	var obj models.UserStock

	result := db.Model(models.UserStock{}).
		Joins("INNER JOIN stocks on stocks.id = user_stocks.stock_id").
		Where("stock_id = ? AND user_id = ? AND account_id = ?", stockID, uid, accountID).
		First(&obj)
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
	uids := auth.GetAllowedUsers(user, true, false, false)
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
	qry.Find(&users)
	return users
}

func GetSnapshot(db *gorm.DB, id uint) (models.StockSnapshot, error) {
	var obj models.StockSnapshot
	result := db.Model(models.StockSnapshot{}).Where("id = ?", id).First(&obj)
	return obj, result.Error
}

func GetVisibleAccounts(db *gorm.DB, user models.User, permitLimited bool) ([]models.Account, error) {
	var accounts []models.Account
	qry := db.Model(&models.Account{})
	if !user.IsAdmin {
		uids := auth.GetAllowedUsers(user, true, false, permitLimited)
		qry = qry.Where("user_id IN ?", uids)
	}
	res := qry.Find(&accounts)
	return accounts, res.Error
}
func GetAccount(db *gorm.DB, id uint) (models.Account, error) {
	var account models.Account
	res := db.Model(&models.Account{}).Where("id = ?", id).First(&account)
	return account, res.Error
}
func GetStocksForAccount(db *gorm.DB, accountID uint) ([]models.UserStock, error) {
	var stocks []models.UserStock
	res := db.Model(&models.UserStock{}).
		Where("account_id = ? AND currently_held = true", accountID).
		Find(&stocks)
	return stocks, res.Error
}

func GetDemoUser(db *gorm.DB) models.User {
	var user models.User
	db.Model(&models.User{}).Where("is_demo_user = TRUE").First(&user)
	return user
}
