package xgin

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"

	"github.com/laixhe/gonet/xgin/constant"
)

// SetRequestID 设置请求ID
func SetRequestID() gin.HandlerFunc {
	return requestid.New(requestid.WithGenerator(func() string {
		return xid.New().String()
	}))
}

// GetRequestID 获取请求ID
func GetRequestID(c *gin.Context) string {
	return requestid.Get(c)
}

// GetPlatform 获取平台
func GetPlatform(c *gin.Context) string {
	return c.Request.Header.Get(constant.HeaderPlatform)
}
