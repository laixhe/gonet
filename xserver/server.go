package xserver

import (
	"errors"
	"fmt"

	"github.com/laixhe/gonet/proto/gen/config/cserver"
	"github.com/laixhe/gonet/xlog"
)

// Checking 检查配置
func Checking(c *cserver.Server) error {
	if c == nil {
		return errors.New("server config is nil")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("server config port error: 1~65535")
	}
	xlog.Debugf("server config=%v", c)
	return nil
}

// AddrIPv4 地址( 0.0.0.0:80 )
func AddrIPv4(c *cserver.Server) string {
	return fmt.Sprintf(":%d", c.Port)
}
