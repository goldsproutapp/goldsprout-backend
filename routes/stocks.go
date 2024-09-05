package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/middleware"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/request"
	"github.com/shopspring/decimal"
)

func GetAllStocks(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	userStocks := database.GetVisibleStockList(user, db)
	ctx.JSON(http.StatusOK, userStocks)
}

func updateHoldingMap(m *map[uint]map[uint]models.HoldingInfo, a uint, b uint, info models.HoldingInfo) {
	_, ok := (*m)[a]
	if !ok {
		(*m)[a] = map[uint]models.HoldingInfo{}
	}
	_, ok = (*m)[a][b]
	if !ok {
		(*m)[a][b] = info
	} else {
		(*m)[a][b] = (*m)[a][b].Merge(info)
	}
}

func GetHoldings(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	userStocks := database.GetVisibleStockList(user, db)
	snapshots := database.GetLatestSnapshots(userStocks, db)
	byUser := map[uint]map[uint]models.HoldingInfo{}
	byStock := map[uint]map[uint]models.HoldingInfo{}
	byAccount := map[uint]map[uint]models.HoldingInfo{}
	for i, snapshot := range snapshots {
		if snapshot == nil {
			continue
		}
		info := models.HoldingInfo{
			Value: decimal.NewFromInt(0),
			Units: decimal.NewFromInt(0),
		}
		if userStocks[i].CurrentlyHeld {
			info.Value = snapshot.Value
			info.Units = snapshot.Units
		}
		updateHoldingMap(&byUser, snapshot.UserID, snapshot.StockID, info)
		updateHoldingMap(&byStock, snapshot.StockID, snapshot.UserID, info)
		updateHoldingMap(&byAccount, snapshot.AccountID, snapshot.StockID, info)
	}
	request.OK(ctx, gin.H{
		"by_user": byUser, "by_stock": byStock, "by_account": byAccount,
	})

}

func UpdateStock(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	var body models.StockUpdateRequest
	err := ctx.BindJSON(&body)
	if err != nil {
		request.BadRequest(ctx)
		return
	}
	if !database.CanModifyStock(db, user, body.Stock.ID) {
		request.Forbidden(ctx)
		return
	}
	db.Save(&(body.Stock))

}

func MergeStocks(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	var body models.StockMergeRequest
	err := ctx.BindJSON(&body)
	if err != nil {
		request.BadRequest(ctx)
		return
	}
	if !database.CanModifyStock(db, user, body.Stock) || !database.CanModifyStock(db, user, body.MergeInto) {
		request.Forbidden(ctx)
		return
	}
	db.Model(&models.StockSnapshot{}).Where("stock_id = ?", body.Stock).Update("stock_id", body.MergeInto)
	var userStocks []models.UserStock
	db.Model(&models.UserStock{}).Where("stock_id = ?", body.Stock).Find(&userStocks)
	for _, us := range userStocks {
		var otherUS models.UserStock
		if !database.Exists(db.Model(&models.UserStock{}).
			Where("stock_id = ? AND user_id = ?", body.MergeInto, us.UserID).
			First(&otherUS)) {
			us.StockID = body.MergeInto
			db.Save(&us)
		} else {
			if us.CurrentlyHeld && !otherUS.CurrentlyHeld {
				otherUS.CurrentlyHeld = true
				db.Save(&otherUS)
			}
			db.Delete(&us)
		}
	}
	db.Model(&models.UserStock{}).Where("stock_id = ?", body.Stock).Update("stock_id", body.MergeInto)
	db.Delete(&models.Stock{}, body.Stock)
	request.NoContent(ctx)
}

func RegisterStockRoutes(router *gin.RouterGroup) {
	router.GET("/holdings", middleware.Authenticate("AccessPermissions"), GetHoldings)
	router.GET("/stocks", middleware.Authenticate("AccessPermissions"), GetAllStocks)
	router.PUT("/stocks", middleware.Authenticate("AccessPermissions"), UpdateStock)
	router.POST("/stocks/merge", middleware.Authenticate("AccessPermissions"), MergeStocks)
}
