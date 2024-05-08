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
					AnnualFee:        0,
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

	objs := []models.StockSnapshot{}
bodyLoop:
	for i, snapshot := range body.Entries {
		userStock := userStocks[i]
		prevSnapshot := prevSnapshots[i]

		errList := []error{}
		price := util.ParseDecimal(snapshot.Price, &errList)
		value := util.ParseDecimal(snapshot.Value, &errList)
		totalChange := util.ParseDecimal(snapshot.AbsoluteChange, &[]error{})
		date := time.Unix(body.Date, 0)

		obj := models.StockSnapshot{
			UserID:                body.UserID,
			Date:                  date,
			StockID:               userStock.StockID,
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

		for j, other := range objs {
			if j >= i {
				break
			}
			if other.StockID == userStock.StockID {
				// Price and normalised performance should be the same in both instances.
				// If it's not, then the input data is bad.
				newOther := models.StockSnapshot{
					UserID:                other.UserID,
					Date:                  other.Date,
					StockID:               other.StockID,
					Units:                 other.Units.Add(obj.Units),
					Price:                 other.Price.Add(obj.Price),
					Cost:                  other.Cost.Add(obj.Cost),
					Value:                 other.Value.Add(obj.Value),
					ChangeToDate:          other.ChangeToDate.Add(obj.ChangeToDate),
					ChangeSinceLast:       calculations.CalculateValueChange(obj.Value.Add(other.Value), prevSnapshot),
					NormalisedPerformance: other.NormalisedPerformance,
				}
				objs[j] = newOther
				continue bodyLoop
			}
		}
		objs = append(objs, obj)
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

func DeleteSnapshot(ctx *gin.Context) {
	errs := []error{}
	id := util.ParseUint(ctx.Param("id"), &errs)
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	snapshot, err := database.GetSnapshot(db, id)
	if err != nil {
		request.NotFound(ctx)
		return
	}
	if !auth.HasAccessPerm(user, snapshot.UserID, false, true) {
		request.Forbidden(ctx)
		return
	}
	db.Delete(&snapshot)
	request.NoContent(ctx)
}

func RegisterSnapshotRoutes(router *gin.RouterGroup) {
	router.GET("/snapshots/latest", middleware.Authenticate("AccessPermissions"), GetLatestSnapshotList)
	router.GET("/snapshots/all", middleware.Authenticate("AccessPermissions"), GetAllSnapshots)
	router.POST("/snapshots", middleware.Authenticate("AccessPermissions"), CreateSnapshots)
	router.DELETE("/snapshots/:id", middleware.Authenticate("AccessPermissions"), DeleteSnapshot)
}
