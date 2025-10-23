package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/auth"
	"github.com/goldsproutapp/goldsprout-backend/database"
	"github.com/goldsproutapp/goldsprout-backend/lib/snapshots"
	"github.com/goldsproutapp/goldsprout-backend/middleware"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/request/response"
	"github.com/goldsproutapp/goldsprout-backend/util"
)

func GetLatestSnapshotList(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	stocks := database.GetHeldStocks(user, db, false)
	all_snapshots := database.GetLatestSnapshots(stocks, db)
	snapshots := []models.StockSnapshot{}
	for _, snapshot := range all_snapshots {
		if snapshot != nil {
			snapshots = append(snapshots, *snapshot)
		}
	}
	response.OK(ctx, snapshots)
}

func GetSnapshotForStock(ctx *gin.Context) {
	idstr, exists := ctx.GetQuery("id")
	if !exists {
		response.BadRequest(ctx)
		return
	}
	id := util.ParseIntOrDefault(idstr, -1)
	if id == -1 {
		response.BadRequest(ctx)
		return
	}
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	users := []uint{}
	if !user.IsAdmin {
		users = auth.GetAllowedUsers(user, true, false, false)
	}
	snapshots := database.GetSnapshots(users, []uint{uint(id)}, db)
	response.OK(ctx, snapshots)
}

func CreateSnapshots(ctx *gin.Context) {
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)

	var body models.StockSnapshotCreationRequest

	err := ctx.BindJSON(&body)
	if err != nil {
		response.BadRequest(ctx)
		return
	}
	out, err := snapshots.CreateSnapshots(user, db, body)
	if err != nil {
		response.SendError(ctx, err)
	}
	response.Created(ctx, out)
}

func DeleteSnapshot(ctx *gin.Context) {
	errs := []error{}
	id := util.ParseUint(ctx.Param("id"), &errs)
	if len(errs) > 0 {
		response.BadRequest(ctx)
		return
	}
	db := middleware.GetDB(ctx)
	user := middleware.GetUser(ctx)
	snapshot, err := database.GetSnapshot(db, id)
	if err != nil {
		response.NotFound(ctx)
		return
	}
	if !auth.HasAccessPerm(user, snapshot.UserID, false, true, false) {
		response.Forbidden(ctx)
		return
	}
	db.Delete(&snapshot)
	response.NoContent(ctx)
}

func RegisterSnapshotRoutes(router *gin.RouterGroup) {
	router.GET("/snapshots/latest", middleware.Authenticate("AccessPermissions"), GetLatestSnapshotList)
	router.GET("/snapshots/for_stock", middleware.Authenticate("AccessPermissions"), GetSnapshotForStock)
	router.POST("/snapshots", middleware.Authenticate("AccessPermissions"), CreateSnapshots)
	router.DELETE("/snapshots/:id", middleware.Authenticate("AccessPermissions"), DeleteSnapshot)
}
