package request

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Redirect(ctx *gin.Context, location string) {
	if ctx.GetHeader("Accept") == "application/json" {
		ctx.JSON(http.StatusOK, gin.H{"type": "link", "redirect": location})
		return
	}
	ctx.Redirect(http.StatusMovedPermanently, location)
}
