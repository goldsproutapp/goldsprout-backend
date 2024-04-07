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
