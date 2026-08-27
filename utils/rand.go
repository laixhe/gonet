package utils

import (
	"math"

	crand "crypto/rand"
	mrand "math/rand/v2"
)

// RandBool 返回随机布尔值
func RandBool() bool {
	return mrand.IntN(2) == 1
}

// RandRange 返回 [min, max] 区间内的随机整数
// 若 min == max 直接返回 min；若 max < min 自动交换两者
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

// RandBytes 生成 n 个加密安全的随机字节
// 优先使用 crypto/rand，失败时回退到 math/rand
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

// RandNumeral 生成 n 位随机数字字符串
func RandNumeral(n int) string {
	if n <= 0 {
		return ""
	}
	data := make([]byte, n)
	for i := range data {
		data[i] = byte(mrand.IntN(10) + '0')
	}
	return string(data)
}

// RandString 生成 n 位随机字符串
// 不带参数：包含大小写字母和数字
// isUpper[0]=true：大写字母 + 数字
// isUpper[0]=false：小写字母 + 数字
func RandString(n int, isUpper ...bool) string {
	if n <= 0 {
		return ""
	}
	data := make([]byte, n)
	if len(isUpper) == 0 {
		// 混合：大写 + 小写 + 数字
		for i := range data {
			switch mrand.IntN(3) {
			case 1:
				data[i] = byte(mrand.IntN(26) + 'a')
			case 2:
				data[i] = byte(mrand.IntN(26) + 'A')
			default:
				data[i] = byte(mrand.IntN(10) + '0')
			}
		}
		return string(data)
	}
	// 指定大小写 + 数字
	letterBase := int('a')
	if isUpper[0] {
		letterBase = int('A')
	}
	for i := range data {
		if mrand.IntN(2) == 1 {
			data[i] = byte(mrand.IntN(26) + letterBase)
		} else {
			data[i] = byte(mrand.IntN(10) + '0')
		}
	}
	return string(data)
}

// RandLetter 生成 n 位随机字母字符串
// 不带参数：混合大小写
// isUpper[0]=true：全大写
// isUpper[0]=false：全小写
func RandLetter(n int, isUpper ...bool) string {
	if n <= 0 {
		return ""
	}
	data := make([]byte, n)
	if len(isUpper) == 0 {
		// 混合大小写
		for i := range data {
			if mrand.IntN(2) == 1 {
				data[i] = byte(mrand.IntN(26) + 'A')
			} else {
				data[i] = byte(mrand.IntN(26) + 'a')
			}
		}
		return string(data)
	}
	// 指定大小写
	letterBase := int('a')
	if isUpper[0] {
		letterBase = int('A')
	}
	for i := range data {
		data[i] = byte(mrand.IntN(26) + letterBase)
	}
	return string(data)
}
