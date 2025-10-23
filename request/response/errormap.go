package response

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/goldsproutapp/goldsprout-backend/lib/exceptions"
)

var errorToResponseFunctionMap = map[error]func(ctx *gin.Context){
	exceptions.ConflictBase:       Conflict,
	exceptions.InvalidRequestBase: BadRequest,
	exceptions.UserForbiddenBase:  Forbidden,
}

func SendError(ctx *gin.Context, err error) {
	for k, v := range errorToResponseFunctionMap {
		if errors.Is(err, k) {
			v(ctx)
			return
		}
	}
	BadRequest(ctx)

}
