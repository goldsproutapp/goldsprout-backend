package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/auth"
	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/middleware"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/request"
	"github.com/goldsproutapp/goldsprout-backend/util"
	"github.com/shopspring/decimal"
)

func GetAccounts(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	accounts, err := database.GetVisibleAccounts(db, user)
	if err != nil {
		request.BadRequest(ctx)
		return
	}
	out := make([]models.AccountReponse, len(accounts))
	for i, acc := range accounts {
		userStocks, err := database.GetStocksForAccount(db, acc.ID)
		if err != nil {
			request.BadRequest(ctx) // TODO: this isn't really the right response code here
			return
		}
		var numStocks uint = 0
		value := decimal.NewFromInt(0)
		snapshots := database.GetLatestSnapshots(userStocks, db)
		for _, snapshot := range snapshots {
			if snapshot != nil {
				value = value.Add(snapshot.Value)
				numStocks += 1
			}
		}
		out[i] = models.AccountReponse{
			Account:    acc,
			Value:      value,
			StockCount: numStocks,
		}
	}
	request.OK(ctx, out)
}

func CreateAccount(ctx *gin.Context) {
	var body models.CreateAccountRequest
	if ctx.BindJSON(&body) != nil {
		request.BadRequest(ctx)
		return
	}
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	if !auth.HasAccessPerm(user, body.UserID, false, true) {
		request.Forbidden(ctx)
		return
	}
	account := models.Account{
		Name:       body.Name,
		ProviderID: body.ProviderID,
		UserID:     body.UserID,
	}
	res := db.Create(&account)
	if res.Error != nil {
		request.BadRequest(ctx)
		return
	}
	request.Created(ctx, account)
}

func DeleteAccount(ctx *gin.Context) {
	errs := []error{}
	id := util.ParseUint(ctx.Param("id"), &errs)
	if len(errs) > 0 {
		request.BadRequest(ctx)
		return
	}
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	account, err := database.GetAccount(db, id)
	if err != nil {
		request.NotFound(ctx)
		return
	}
	if !auth.HasAccessPerm(user, account.UserID, true, true) {
		request.Forbidden(ctx)
		return
	}
	db.Where("account_id = ?", account.ID).Delete(&models.StockSnapshot{})
	db.Where("account_id = ?", account.ID).Delete(&models.UserStock{})
	db.Delete(&models.Account{}, account.ID)
	request.NoContent(ctx)
}

func RegisterAccountRoutes(router *gin.RouterGroup) {
	router.GET("/accounts", middleware.Authenticate("AccessPermissions"), GetAccounts)
	router.POST("/accounts", middleware.Authenticate("AccessPermissions"), CreateAccount)
	router.DELETE("/accounts/:id", middleware.Authenticate("AccessPermissions"), DeleteAccount)
}
