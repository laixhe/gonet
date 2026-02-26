package utils

import (
	"crypto/rand"
	"io"
	"math/big"
)

// RandInt64 生成区间随机数
func RandInt64(min, max int64) int64 {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	result, _ := rand.Int(rand.Reader, big.NewInt(max-min))
	data := result.Int64() + min
	return data
}

// RandBytes 随机生成 n 个字节
func RandBytes(n int) ([]byte, error) {
	data := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		return nil, err
	}
	return data, nil
}
