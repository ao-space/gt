package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	SUCCESS = 200
	ERROR   = 500
	OVERDUE = 401
)
const SuccessMsg = "SUCCESS"
const ErrorMsg = "FAIL"
const OverdueMsg = "Token is overdue"
const InvalidKeyMsg = "Invalid key"

func Response(httpStatus int, code int, data interface{}, msg string, ctx *gin.Context) {
	ctx.JSON(httpStatus, gin.H{
		"code": code,
		"data": data,
		"msg":  msg,
	})
}
func Success(ctx *gin.Context) {
	Response(http.StatusOK, SUCCESS, nil, SuccessMsg, ctx)
}
func SuccessWithMessage(message string, ctx *gin.Context) {
	Response(http.StatusOK, SUCCESS, nil, message, ctx)
}

func SuccessWithData(data interface{}, ctx *gin.Context) {
	Response(http.StatusOK, SUCCESS, data, SuccessMsg, ctx)
}
func SuccessWithDetailed(data interface{}, message string, ctx *gin.Context) {
	Response(http.StatusOK, SUCCESS, data, message, ctx)
}

func Fail(ctx *gin.Context) {
	Response(http.StatusOK, ERROR, nil, ErrorMsg, ctx)
}
func FailWithMessage(message string, ctx *gin.Context) {

	Response(http.StatusOK, ERROR, nil, message, ctx)
}
func FailWithDetailed(data interface{}, message string, ctx *gin.Context) {
	Response(http.StatusOK, ERROR, data, message, ctx)
}
func InvalidToken(ctx *gin.Context) {
	Response(http.StatusOK, OVERDUE, nil, OverdueMsg, ctx)
}

func InvalidKey(ctx *gin.Context) {
	Response(http.StatusOK, OVERDUE, nil, InvalidKeyMsg, ctx)
}
