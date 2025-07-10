package orm

import "fmt"

// PageLimitOffset 分页
// page: 当前页
// pageSize: 每页数量
func PageLimitOffset[T int | int32 | int64](page, pageSize T) (limit T, offset T) {
	return pageSize, (page - 1) * pageSize
}

// PageLimitOffsetSql 分页SQL
// limit:  数量
// offset: 偏移数量
func PageLimitOffsetSql[T int | int32 | int64](limit, offset T) string {
	if offset == 0 {
		return fmt.Sprintf("LIMIT %d", limit)
	}
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}
