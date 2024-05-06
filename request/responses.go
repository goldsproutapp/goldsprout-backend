package request

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Created(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusCreated, gin.H{"success": true, "message": "created", "data": data})
}

func OK(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func NoContent(ctx *gin.Context) {
	ctx.Status(http.StatusNoContent)
}
