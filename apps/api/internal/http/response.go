package http

import (
	stdhttp "net/http"

	"github.com/gin-gonic/gin"
)

func OK(ctx *gin.Context, data any) {
	ctx.JSON(stdhttp.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    data,
	})
}

func Fail(ctx *gin.Context, status int, code int, message string) {
	ctx.JSON(status, gin.H{
		"code":    code,
		"message": message,
		"data":    nil,
	})
}
