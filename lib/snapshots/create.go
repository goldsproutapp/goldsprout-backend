package snapshots

import (
	"strconv"
	"time"

	"github.com/goldsproutapp/goldsprout-backend/auth"
	"github.com/goldsproutapp/goldsprout-backend/calculations"
	"github.com/goldsproutapp/goldsprout-backend/constants"
	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/lib/exceptions"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func CreateSnapshots(user models.User, db *gorm.DB, request models.StockSnapshotCreationRequest) ([]models.StockSnapshot, error) {
	out := []models.StockSnapshot{}
	for _, batch := range request.Batches {
		account, err := database.GetAccount(db, batch.AccountID)
		if err != nil {
			return nil, exceptions.InvalidRequest("")
		}
		if !auth.HasAccessPerm(user, account.UserID, false, true, false) {
			return nil, exceptions.UserForbidden("")
		}
		userStocks := make([]models.UserStock, len(batch.Entries))
		stockIDs := util.NewHashSet[uint]()
		providerIDs := util.NewHashSet[uint]()
		for i, snapshot := range batch.Entries {
			globalStock, err := database.GetGlobalStockByNameOrCode(
				db, snapshot.StockName, snapshot.StockCode, account.ProviderID)
			if err != nil {
				globalStock = models.Stock{
					Name:                snapshot.StockName,
					ProviderID:          account.ProviderID,
					Sector:              constants.DEFAULT_SECTOR_NAME,
					Region:              constants.DEFAULT_REGION_NAME,
					StockCode:           snapshot.StockCode,
					NeedsAttention:      true, // The defaults set above need manually reviewing
					TrackingStrategy:    constants.STRATEGY_DATA_IMPORT,
					AnnualFee:           0,
					ClassCompositionMap: map[string]decimal.Decimal{constants.DEFAULT_CLASS_NAME: decimal.NewFromInt(100)},
				}
				res := db.Create(&globalStock)
				if res.Error != nil {
					return nil, exceptions.InvalidRequest("")
				}
			}
			userStock, err := database.GetUserStock(db, account.UserID, globalStock.ID, account.ID)
			if err != nil {
				userStock = models.UserStock{
					UserID:        account.UserID,
					StockID:       globalStock.ID,
					AccountID:     account.ID,
					CurrentlyHeld: true,
					Notes:         "",
				}
				res := db.Create(&userStock)
				if res.Error != nil {
					return nil, exceptions.InvalidRequest("")
				}
			}
			// NOTE: is there a case to made for relaxing the permission requirements here?
			if database.CanModifyStock(db, user, globalStock.ID) {
				globalStock.Region = util.UpdateIfSet(globalStock.Region, snapshot.Region)
				globalStock.Sector = util.UpdateIfSet(globalStock.Sector, snapshot.Sector)
				if snapshot.AnnualFee != "" {
					fee, err := strconv.ParseFloat(snapshot.AnnualFee, 32)
					if err == nil {
						globalStock.AnnualFee = float32(fee)
					}
				}
				globalStock.Name = snapshot.StockName
				if snapshot.StockCode != "" {
					globalStock.StockCode = snapshot.StockCode
				}
				db.Save(&globalStock)
			}

			if !userStock.CurrentlyHeld {
				userStock.CurrentlyHeld = true
				db.Save(&userStock)
			}
			userStocks[i] = userStock
			stockIDs.Add(userStock.StockID)
			providerIDs.Add(account.ProviderID)
		}
		prevSnapshots := database.GetLatestSnapshots(userStocks, db)

		objs := []models.StockSnapshot{}
	bodyLoop:
		for i, snapshot := range batch.Entries {
			userStock := userStocks[i]
			prevSnapshot := prevSnapshots[i]

			date := time.Unix(batch.Date, 0)
			if prevSnapshot != nil && date.Sub(prevSnapshot.Date).Abs().Hours() < 1 {
				return nil, exceptions.Conflict("")
			}

			errList := []error{}
			price := util.ParseDecimal(snapshot.Price, &errList)
			value := util.ParseDecimal(snapshot.Value, &errList)
			totalChange := util.ParseDecimal(snapshot.AbsoluteChange, &[]error{})
			cost := util.ParseDecimal(snapshot.Cost, &errList)
			units := util.ParseDecimal(snapshot.Units, &errList)

			obj := models.StockSnapshot{
				UserID:                 account.UserID,
				Date:                   date,
				StockID:                userStock.StockID,
				AccountID:              account.ID,
				Units:                  units,
				Price:                  price,
				Cost:                   cost,
				Value:                  value,
				ChangeToDate:           totalChange,
				ChangeSinceLast:        calculations.CalculateValueChange(totalChange, prevSnapshot),
				NormalisedPerformance:  calculations.CalculateNormalisedPerformance(price, prevSnapshot, date),
				TransactionAttribution: snapshot.TransactionAttribution,
			}
			if len(errList) != 0 {
				return nil, exceptions.InvalidRequest("")
			}

			for j, other := range objs {
				if j >= i {
					break
				}
				if other.StockID == userStock.StockID {
					// Price and normalised performance should be the same in both instances.
					// If it's not, then the input data is bad.
					newOther := models.StockSnapshot{
						UserID:                 other.UserID,
						Date:                   other.Date,
						StockID:                other.StockID,
						AccountID:              other.AccountID,
						Units:                  other.Units.Add(obj.Units),
						Price:                  other.Price.Add(obj.Price),
						Cost:                   other.Cost.Add(obj.Cost),
						Value:                  other.Value.Add(obj.Value),
						ChangeToDate:           other.ChangeToDate.Add(obj.ChangeToDate),
						ChangeSinceLast:        calculations.CalculateValueChange(obj.ChangeToDate.Add(other.ChangeToDate), prevSnapshot),
						NormalisedPerformance:  other.NormalisedPerformance,
						TransactionAttribution: other.TransactionAttribution,
					}
					objs[j] = newOther
					continue bodyLoop
				}
			}
			objs = append(objs, obj)
		}
		if len(objs) > 0 {
			result := db.Create(&objs)
			if result.Error != nil {
				return nil, exceptions.InvalidRequest("")
			}
		}
		if batch.DeleteSoldStocks {
			var toUpdate []models.UserStock
			qry := db.Model(&models.UserStock{}).
				Where("account_id = ?", account.ID)
			if stockIDs.Size() > 0 {
				qry = qry.Where("stock_id NOT IN ?", stockIDs.Items())
			}
			qry.Find(&toUpdate)
			toUpdateIDs := make([]uint, len(toUpdate))
			for i, us := range toUpdate {
				toUpdateIDs[i] = us.ID
			}
			db.Model(&models.UserStock{}).Where("id IN ?", toUpdateIDs).Update("currently_held", false)
		}
		out = append(out, objs...)
	}
	return out, nil
}
