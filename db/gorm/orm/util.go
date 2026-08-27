package orm

import (
	"fmt"

	"gorm.io/gorm"
)

func Model(db *gorm.DB, model any) *gorm.DB {
	if s, ok := model.(string); ok {
		return db.Table(s)
	}
	return db.Model(model)
}

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

// maxIntegerValue 返回泛型整数的最大值
func maxIntegerValue[T int | int32 | int64]() T {
	switch any(T(0)).(type) {
	case int32:
		// 先赋值给具体类型变量, 避免常量直接转换 int32 溢出
		v := int32(^uint32(0) >> 1)
		return T(v)
	case int64:
		v := int64(^uint64(0) >> 1)
		return T(v)
	default:
		v := int(^uint(0) >> 1)
		return T(v)
	}
}

// PageOffsetCalculation 分页数量计算
// page: 当前页
// pageSize: 每页数量
func PageOffsetCalculation[T int | int32 | int64](page, pageSize T) (limit T, offset T) {
	limit = pageSize
	if page < 1 {
		page = 1
	}
	// 防止偏移量 (page-1)*pageSize 溢出, 页码过大时截断为不会溢出的最大页码
	if pageSize > 0 && page-1 > maxIntegerValue[T]()/pageSize {
		page = maxIntegerValue[T]()/pageSize + 1
	}
	return limit, (page - 1) * pageSize
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
