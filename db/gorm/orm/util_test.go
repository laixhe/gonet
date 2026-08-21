package orm

import (
	"math"
	"testing"
)

func TestPageParamCheck(t *testing.T) {
	tests := []struct {
		name              string
		page, pageSize    int
		pageSizeDefault   int
		wantPage, wantSize int
	}{
		{"零值使用默认值", 0, 0, 10, 1, 10},
		{"page 为 0 pageSize 有效", 0, 20, 10, 1, 20},
		{"page 有效 pageSize 为 0", 5, 0, 10, 5, 10},
		{"page 为负数", -3, 20, 10, 1, 20},
		{"全部有效", 5, 20, 10, 5, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, s := PageParamCheck(tt.page, tt.pageSize, tt.pageSizeDefault)
			if p != tt.wantPage || s != tt.wantSize {
				t.Errorf("PageParamCheck(%d, %d, %d) = (%d, %d), want (%d, %d)",
					tt.page, tt.pageSize, tt.pageSizeDefault, p, s, tt.wantPage, tt.wantSize)
			}
		})
	}
}

func TestPageOffsetCalculation(t *testing.T) {
	t.Run("普通分页", func(t *testing.T) {
		if limit, offset := PageOffsetCalculation(1, 10); limit != 10 || offset != 0 {
			t.Errorf("PageOffsetCalculation(1, 10) = (%d, %d), want (10, 0)", limit, offset)
		}
		if limit, offset := PageOffsetCalculation(2, 10); limit != 10 || offset != 10 {
			t.Errorf("PageOffsetCalculation(2, 10) = (%d, %d), want (10, 10)", limit, offset)
		}
	})
	t.Run("page 小于 1 归一为 1", func(t *testing.T) {
		if limit, offset := PageOffsetCalculation(0, 10); limit != 10 || offset != 0 {
			t.Errorf("PageOffsetCalculation(0, 10) = (%d, %d), want (10, 0)", limit, offset)
		}
	})
	t.Run("int64 溢出防护", func(t *testing.T) {
		var page, pageSize int64 = math.MaxInt64, 10
		_, offset := PageOffsetCalculation(page, pageSize)
		want := int64(math.MaxInt64/10) * 10
		if offset != want {
			t.Errorf("offset = %d, want %d", offset, want)
		}
	})
	t.Run("int32 溢出防护", func(t *testing.T) {
		var page, pageSize int32 = math.MaxInt32, 100
		_, offset := PageOffsetCalculation(page, pageSize)
		want := int32(math.MaxInt32/100) * 100
		if offset != want {
			t.Errorf("offset = %d, want %d", offset, want)
		}
	})
}

func TestPageOffsetSql(t *testing.T) {
	if s := PageOffsetSql(10, 0); s != "LIMIT 10" {
		t.Errorf("PageOffsetSql(10, 0) = %q, want %q", s, "LIMIT 10")
	}
	if s := PageOffsetSql(10, 20); s != "LIMIT 10 OFFSET 20" {
		t.Errorf("PageOffsetSql(10, 20) = %q, want %q", s, "LIMIT 10 OFFSET 20")
	}
}

func TestModel(t *testing.T) {
	t.Run("字符串表名", func(t *testing.T) {
		db := newTestDB(t)
		tx := Model(db, "users")
		if tx.Statement.Table != "users" {
			t.Errorf("Table = %q, want %q", tx.Statement.Table, "users")
		}
	})
	t.Run("结构体模型", func(t *testing.T) {
		db := newTestDB(t)
		tx := Model(db, &testModel{})
		if tx.Statement.Model == nil {
			t.Error("Model 不应为 nil")
		}
	})
}

func TestMaxIntegerValue(t *testing.T) {
	if v := maxIntegerValue[int32](); v != math.MaxInt32 {
		t.Errorf("maxIntegerValue[int32]() = %d, want %d", v, int32(math.MaxInt32))
	}
	if v := maxIntegerValue[int64](); v != math.MaxInt64 {
		t.Errorf("maxIntegerValue[int64]() = %d, want %d", v, int64(math.MaxInt64))
	}
	if v := maxIntegerValue[int](); v != int(^uint(0)>>1) {
		t.Errorf("maxIntegerValue[int]() = %d, want %d", v, int(^uint(0)>>1))
	}
}
