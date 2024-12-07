package xgorm

import "fmt"

// PageLimit 分页
func PageLimit(page, pageSize int) (limit int, offset int) {
	return pageSize, (page - 1) * pageSize
}

// PageLimitSql 分页SQL
func PageLimitSql(offset, limit int) string {
	if offset == 0 {
		return fmt.Sprintf("LIMIT %d", limit)
	}
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}
