package utils

import (
	"math"
	"sync"
	"time"

	crand "crypto/rand"
	mrand "math/rand/v2"
)

// RandPool 为每个 goroutine 独立的 rand.Rand，避免并发问题
var RandPool = sync.Pool{
	New: func() any {
		return mrand.New(mrand.NewPCG(uint64(time.Now().UnixNano()), 0))
	},
}

// RandBool 随机布尔值
func RandBool() bool {
	return mrand.IntN(2) == 1
}

// RandRange 生成区间随机数
// min 不能小于 0
// max 不能小于等于 0
func RandRange(min, max int) int {
	if min == max {
		return min
	}
	if max < min {
		min, max = max, min
	}
	if min == 0 && max == math.MaxInt {
		return mrand.Int()
	}
	return mrand.IntN(max-min+1) + min
}

// RandBytes 随机生成 n 个字节
func RandBytes(n int) []byte {
	if n <= 0 {
		return []byte{}
	}
	data := make([]byte, n)
	if _, err := crand.Read(data); err == nil {
		return data
	}
	for k := range data {
		data[k] = byte(mrand.IntN(256))
	}
	return data
}

// RandNumeral 随机生成整数字符串
func RandNumeral(n int) string {
	if n <= 0 {
		return ""
	}
	data := make([]byte, n)
	for i := range data {
		data[i] = byte(mrand.IntN(10) + 48)
	}
	return string(data)
}

// RandString 随机生成字符串，默认包含大小写字母和整数
func RandString(n int, isUpper ...bool) string {
	if n <= 0 {
		return ""
	}
	data := make([]byte, n)
	if len(isUpper) == 0 {
		for i := range data {
			randInt := mrand.IntN(3)
			switch randInt {
			case 1:
				data[i] = byte(mrand.IntN(26) + 97)
			case 2:
				data[i] = byte(mrand.IntN(26) + 65)
			default:
				data[i] = byte(mrand.IntN(10) + 48)
			}
		}
		return string(data)
	}
	// 区分大小写与整数
	for i := range data {
		if mrand.IntN(2) == 1 {
			if isUpper[0] {
				data[i] = byte(mrand.IntN(26) + 65)
			} else {
				data[i] = byte(mrand.IntN(26) + 97)
			}
		} else {
			data[i] = byte(mrand.IntN(10) + 48)
		}
	}
	return string(data)
}

// RandLetter 随机生成字母字符串，默认包含不区分大小写
// isUpper 可选：是否大写
func RandLetter(n int, isUpper ...bool) string {
	if n <= 0 {
		return ""
	}
	data := make([]byte, n)
	// 不区分大小写
	if len(isUpper) == 0 {
		for i := range data {
			if mrand.IntN(2) == 1 {
				data[i] = byte(mrand.IntN(26) + 65)
			} else {
				data[i] = byte(mrand.IntN(26) + 97)
			}
		}
		return string(data)
	}
	// 区分大小写
	for i := range data {
		if isUpper[0] {
			data[i] = byte(mrand.IntN(26) + 65)
		} else {
			data[i] = byte(mrand.IntN(26) + 97)
		}
	}
	return string(data)
}
