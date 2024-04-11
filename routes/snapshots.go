package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker-backend/auth"
	"github.com/patrickjonesuk/investment-tracker-backend/calculations"
	"github.com/patrickjonesuk/investment-tracker-backend/constants"
	"github.com/patrickjonesuk/investment-tracker-backend/database"
	"github.com/patrickjonesuk/investment-tracker-backend/middleware"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/request"
	"github.com/patrickjonesuk/investment-tracker-backend/util"
)

func GetLatestSnapshotList(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	stocks := database.GetHeldStocks(user, db)
	all_snapshots := database.GetLatestSnapshots(stocks, db)
	snapshots := []models.StockSnapshot{}
	for _, snapshot := range all_snapshots {
		if snapshot != nil {
			snapshots = append(snapshots, *snapshot)
		}
	}
	ctx.JSON(http.StatusOK, snapshots)
}

func GetAllSnapshots(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	snapshots := database.GetAllSnapshots(user, db)
	ctx.JSON(http.StatusOK, snapshots)
}

func CreateSnapshots(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)

	var body models.StockSnapshotCreationRequest

	err := ctx.BindJSON(&body)
	if err != nil {
		request.BadRequest(ctx)
		return
	}
	if !auth.HasAccessPerm(user, body.UserID, false, true) {
		request.Forbidden(ctx)
		return
	}
	userStocks := make([]models.UserStock, len(body.Entries))
	stockIDs := make([]uint, len(body.Entries))
	providerIDs := util.NewHashSet[uint]()
	for i, snapshot := range body.Entries {
		userStock, err := database.GetUserStockByName(db, body.UserID, snapshot.StockName)
		if err != nil {
			globalStock, err := database.GetGlobalStockByName(db, snapshot.StockName, snapshot.ProviderID)
			if err != nil {
				globalStock = models.Stock{
					Name:             snapshot.StockName,
					ProviderID:       snapshot.ProviderID,
					Sector:           constants.DEFAULT_SECTOR_NAME,
					Region:           constants.DEFAULT_REGION_NAME,
					StockCode:        snapshot.StockCode,
					NeedsAttention:   true, // The defaults set above need manually reviewing
					TrackingStrategy: constants.STRATEGY_DATA_IMPORT,
				}
				res := db.Create(&globalStock)
				if res.Error != nil {
					request.BadRequest(ctx)
					return
				}
			}
			userStock = models.UserStock{
				UserID:        body.UserID,
				StockID:       globalStock.ID,
				CurrentlyHeld: true,
				Notes:         "",
			}
			res := db.Create(&userStock)
			if res.Error != nil {
				request.BadRequest(ctx)
				return
			}
		}
        if !userStock.CurrentlyHeld {
            userStock.CurrentlyHeld = true
            db.Save(&userStock)
        }
		userStocks[i] = userStock
		stockIDs[i] = userStock.StockID
		providerIDs.Add(snapshot.ProviderID)
	}
	prevSnapshots := database.GetLatestSnapshots(userStocks, db)

	objs := make([]models.StockSnapshot, len(body.Entries))
	for i, snapshot := range body.Entries {
		userStock := userStocks[i]
		prevSnapshot := prevSnapshots[i]

		errList := []error{}
		price := util.ParseDecimal(snapshot.Price, &errList)
		value := util.ParseDecimal(snapshot.Value, &errList)
		totalChange := util.ParseDecimal(snapshot.AbsoluteChange, &errList)
		date := time.Unix(body.Date, 0)
		obj := models.StockSnapshot{
			UserID:                body.UserID,
			Date:                  date,
			StockID:               userStock.ID,
			Units:                 util.ParseDecimal(snapshot.Units, &errList),
			Price:                 price,
			Cost:                  util.ParseDecimal(snapshot.Cost, &errList),
			Value:                 value,
			ChangeToDate:          totalChange,
			ChangeSinceLast:       calculations.CalculateValueChange(value, prevSnapshot),
			NormalisedPerformance: calculations.CalculateNormalisedPerformance(price, prevSnapshot, date),
		}
		if len(errList) != 0 {
			request.BadRequest(ctx)
			return
		}
		objs[i] = obj
	}
	result := db.Create(&objs)
	if result.Error != nil {
		request.BadRequest(ctx)
		return
	}
	if body.DeleteSoldStocks {
		var toUpdate []models.UserStock
		db.Model(&models.UserStock{}).
			Joins("INNER JOIN stocks on stocks.id = user_stocks.stock_id").
			Where("user_id = ?", body.UserID).
			Where("stock_id NOT IN ?", stockIDs).
			Where("stocks.provider_id IN ?", providerIDs.Items()).
			Find(&toUpdate)
		toUpdateIDs := make([]uint, len(toUpdate))
		for i, us := range toUpdate {
			toUpdateIDs[i] = us.ID
		}
		db.Model(&models.UserStock{}).Where("id IN ?", toUpdateIDs).Update("currently_held", false)
	}
	request.Created(ctx, objs)
}

func RegisterSnapshotRoutes(router *gin.RouterGroup) {
	router.GET("/snapshots/latest", middleware.Authenticate("AccessPermissions"), GetLatestSnapshotList)
	router.GET("/snapshots/all", middleware.Authenticate("AccessPermissions"), GetAllSnapshots)

	router.POST("/snapshots", middleware.Authenticate("AccessPermissions"), CreateSnapshots)
}
