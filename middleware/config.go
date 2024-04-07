package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func InjectObject[T any](config T) gin.HandlerFunc {
    key := fmt.Sprintf("%T", config)
    return func(ctx *gin.Context) {
        ctx.Set(key, config)
    }
}

func GetObject[T any](ctx *gin.Context) T {
    key := fmt.Sprintf("%T", *new(T))
    val := ctx.MustGet(key)
    if cfg, ok := val.(T); ok {
        return cfg
    }
    return *new(T)
}
