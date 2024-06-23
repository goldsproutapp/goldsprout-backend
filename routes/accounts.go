package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker-backend/auth"
	"github.com/patrickjonesuk/investment-tracker-backend/database"
	"github.com/patrickjonesuk/investment-tracker-backend/middleware"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/request"
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

func RegisterAccountRoutes(router *gin.RouterGroup) {
	router.GET("/accounts", middleware.Authenticate("AccessPermissions"), GetAccounts)
	router.POST("/accounts", middleware.Authenticate("AccessPermissions"), CreateAccount)
}
