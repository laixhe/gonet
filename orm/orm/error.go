package orm

import (
	"errors"

	"gorm.io/gorm"
)

// ErrorNoUpdatedLines 没有更新行
var ErrorNoUpdatedLines = errors.New("no updated lines")

// IsRecordNotFound 是否未找到记录
func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// IsNoUpdatedLines 是否没有更新行
func IsNoUpdatedLines(err error) bool {
	return errors.Is(err, ErrorNoUpdatedLines)
}
