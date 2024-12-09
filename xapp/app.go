package xapp

import (
	"errors"

	"github.com/laixhe/gonet/protocol/gen/config/capp"
	"github.com/laixhe/gonet/xlog"
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
	xlog.Debugf("app config=%v", c)
	return nil
}
