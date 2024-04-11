package request

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func BadRequest(ctx *gin.Context) {
	ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": "false", "message": "bad request"})
}

func Forbidden(ctx *gin.Context) {
	ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": "false", "message": "forbidden"})
}

func Conflict(ctx *gin.Context) {
    ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"success": "false", "message": "conflict"})
}

func NotFound(ctx *gin.Context) {
    ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"success": "false", "message": "not found"})
}
