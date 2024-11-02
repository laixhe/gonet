package serverx

import (
	"errors"
	"fmt"

	"github.com/laixhe/gonet/logx"
	"github.com/laixhe/gonet/proto/gen/config/cserver"
)

// Checking 检查配置
func Checking(c *cserver.Server) error {
	if c == nil {
		return errors.New("server config is nil")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("server config port error: 1~65535")
	}
	logx.Debugf("server config=%v", c)
	return nil
}

// Addr 地址( 0.0.0.0:80 )
func AddrIPv4(c *cserver.Server) string {
	return fmt.Sprintf(":%d", c.Port)
}
