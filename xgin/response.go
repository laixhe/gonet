package xgin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/laixhe/gonet/xerror"
	"github.com/laixhe/gonet/xresponse"
)

// SuccessJSON 成功
func SuccessJSON(c *gin.Context, data any) {
	c.JSON(http.StatusOK, &xresponse.ResponseModel{
		Data: data,
	})
}

// SuccessJSON200 成功
func SuccessJSON200(c *gin.Context, data any) {
	c.JSON(http.StatusOK, &xresponse.ResponseModel{
		Code: 200,
		Data: data,
	})
}

// ErrorJSON 错误
func ErrorJSON(c *gin.Context, err xerror.IError) {
	c.JSON(http.StatusOK, xresponse.Error(err))
}
