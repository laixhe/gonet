package ginx

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/laixhe/gonet/proto/gen/ecode"
	"github.com/laixhe/gonet/responsex"
)

// SuccessJSON 成功
func SuccessJSON(c *gin.Context, data any) {
	c.JSON(http.StatusOK, &responsex.ResponseModel{
		Code: ecode.ECode_Success,
		Data: data,
	})
}

// ErrorJSON 错误
func ErrorJSON(c *gin.Context, err error) {
	c.JSON(http.StatusOK, responsex.ResponseError(err))
}
