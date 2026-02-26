package orm

import "fmt"

// PageParamCheck 检查分页参数
// page: 当前页
// pageSize: 每页数量
// pageSizeDefault: 每页数量默认值
func PageParamCheck[T int | int32 | int64](page, pageSize, pageSizeDefault T) (T, T) {
	if pageSize <= 0 {
		return max(page, 1), pageSizeDefault
	}
	return max(page, 1), pageSize
}

// PageOffsetCalculation 分页数量计算
// page: 当前页
// pageSize: 每页数量
func PageOffsetCalculation[T int | int32 | int64](page, pageSize T) (limit T, offset T) {
	return pageSize, (page - 1) * pageSize
}

// PageOffsetSql 分页SQL
// limit:  数量
// offset: 偏移数量
func PageOffsetSql[T int | int32 | int64](limit, offset T) string {
	if offset == 0 {
		return fmt.Sprintf("LIMIT %d", limit)
	}
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}
