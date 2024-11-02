package ginx

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/laixhe/gonet/proto/gen/ecode"
	"github.com/laixhe/gonet/responsex"
)

// Success gin返回成功
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, &responsex.ResponseModel{
		Code: ecode.ECode_Success,
		Data: data,
	})
}

// Error gin返回错误
func Error(c *gin.Context, err error) {
	c.JSON(http.StatusOK, responsex.ResponseError(err))
}
