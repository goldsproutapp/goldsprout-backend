package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


const CtxDbKey = "DB"

func Database(db *gorm.DB) gin.HandlerFunc {
    return func(ctx *gin.Context) {
        ctx.Set(CtxDbKey, db)
    }
}

func GetDB(ctx *gin.Context) *gorm.DB {
    val := ctx.MustGet(CtxDbKey)
    if db, ok := val.(*gorm.DB); ok {
        return db
    }
    return nil
}
