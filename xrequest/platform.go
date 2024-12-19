package xrequest

import (
	"github.com/gin-gonic/gin"
)

// 头部 平台

const PlatformHeaderKey = "platform"

func GinPlatformContext(c *gin.Context) string {
	return c.Request.Header.Get(PlatformHeaderKey)
}
