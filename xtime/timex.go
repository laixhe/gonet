package xtime

import "database/sql"

// NullTimeUnixMilli 时间戳毫秒(sql.NullTime 转 int64)
func NullTimeUnixMilli(t sql.NullTime) int64 {
	if t.Valid {
		return t.Time.UnixMilli()
	}
	return 0
}

// NullTimeUnix 时间戳秒(sql.NullTime 转 int64)
func NullTimeUnix(t sql.NullTime) int64 {
	if t.Valid {
		return t.Time.Unix()
	}
	return 0
}
