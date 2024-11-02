package requestx

import (
	"github.com/gin-gonic/gin"

	"github.com/laixhe/gonet/proto/gen/enum/eapp"
)

// 头部 平台

const PlatformHeaderKey = "platform"

func GetPlatform(c *gin.Context) eapp.Platform {
	platform := c.Request.Header.Get(PlatformHeaderKey)
	if platform != "" {
		value := eapp.Platform_value[platform]
		if value != 0 {
			return eapp.Platform(value)
		}
	}
	return eapp.Platform_unknown
}

func IsPlatform(c *gin.Context) bool {
	platform := GetPlatform(c)
	return platform != eapp.Platform_unknown
}
