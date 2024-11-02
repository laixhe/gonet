package appx

import (
	"errors"

	"github.com/laixhe/gonet/logx"
	"github.com/laixhe/gonet/proto/gen/config/capp"
)

// Checking 检查配置
func Checking(c *capp.App) error {
	if c == nil {
		return errors.New("app config is nil")
	}
	if c.Version == "" {
		c.Version = "v0.1"
	}
	if c.Mode == "" {
		c.Mode = capp.ModeType_debug.String()
	} else {
		c.Mode = capp.ModeType_name[capp.ModeType_value[c.Mode]]
	}
	logx.Debugf("app config=%v", c)
	return nil
}
