package routes

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/middleware"
	"github.com/goldsproutapp/goldsprout-backend/request"
)

func GetPreferences(ctx *gin.Context) {
	user := middleware.GetUser(ctx)
	var jsonData interface{}
	err := json.Unmarshal([]byte(user.ClientOpts), &jsonData)
	if err != nil {
		request.OK(ctx, gin.H{})
	} else {
		request.OK(ctx, jsonData)
	}
}

func SetPreferenecs(ctx *gin.Context) {
	var jsonData any
	err := ctx.BindJSON(&jsonData)
	if err != nil {
		request.BadRequest(ctx)
		return
	}
	jsonString, err := json.Marshal(jsonData)
	if err != nil {
		request.BadRequest(ctx)
		return
	}
	user := middleware.GetUser(ctx)
	db := middleware.GetDB(ctx)
	user.ClientOpts = string(jsonString)
	db.Save(&user)
	request.OK(ctx, jsonData)
}

func RegisterPreferencesRoutes(router *gin.RouterGroup) {
	router.GET("/preferences", middleware.Authenticate(), GetPreferences)
	router.PUT("/preferences", middleware.Authenticate(), SetPreferenecs)
}
