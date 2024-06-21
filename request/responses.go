package request

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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

func FileOK(ctx *gin.Context, filename string, content string) {
    ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
    ctx.Header("Content-Type", "text/plain")
    ctx.Writer.WriteString(content)
}
