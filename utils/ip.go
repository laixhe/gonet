package utils

import (
	"net"
)

// IPToInt 把 IPv4 字符串转为 uint32 数值（大端序 / 网络字节序）
// 例如 "192.168.1.1" → 0xC0A80101
func IPToInt(ipaddr string) uint32 {
	ip4 := net.ParseIP(ipaddr).To4()
	if ip4 == nil {
		return 0
	}
	// 大端排列：最高字节 ip4[0] 放在 uint32 的最高位（bit 31-24）
	return uint32(ip4[3]) | uint32(ip4[2])<<8 | uint32(ip4[1])<<16 | uint32(ip4[0])<<24
}

// IPToString 把 uint32 数值转为 IPv4 字符串
func IPToString(ipInt uint32) string {
	return net.IPv4(
		byte(ipInt>>24),
		byte(ipInt>>16&0xFF),
		byte(ipInt>>8&0xFF),
		byte(ipInt&0xFF),
	).String()
}
