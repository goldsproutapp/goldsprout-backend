package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/patrickjonesuk/investment-tracker-backend/auth"
	"github.com/patrickjonesuk/investment-tracker-backend/database"
	"github.com/patrickjonesuk/investment-tracker-backend/middleware"
	"github.com/patrickjonesuk/investment-tracker-backend/models"
	"github.com/patrickjonesuk/investment-tracker-backend/request"
)


func GetAccounts(ctx *gin.Context) {
    db := middleware.GetDB(ctx)
    user := middleware.GetUser(ctx)
    accounts, err := database.GetVisibleAccounts(db, user)
    if err != nil {
        request.BadRequest(ctx)
    } else {
        request.OK(ctx, accounts)
    }
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
        Name: body.Name,
        ProviderID: body.ProviderID,
        UserID: body.UserID,
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
}
