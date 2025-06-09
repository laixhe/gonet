package utils

import (
	"net"
)

// IPToInt 把 ip 字符串转为数值
func IPToInt(ipaddr string) uint32 {
	parseIP := net.ParseIP(ipaddr).To4()
	if parseIP == nil {
		return 0
	}
	ip4 := parseIP.To4()
	if ip4 == nil {
		return 0
	}
	// 数组以大端对齐的方式
	return uint32(ip4[3]) | uint32(ip4[2])<<8 | uint32(ip4[1])<<16 | uint32(ip4[0])<<24
}

// IPToString 把数值转为 ip 字符串
func IPToString(ipInt uint32) string {
	return net.IPv4(
		byte(ipInt>>24),
		byte(ipInt>>16&0xFF),
		byte(ipInt>>8&0xFF),
		byte(ipInt&0xFF),
	).String()
	//return fmt.Sprintf("%d.%d.%d.%d", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}
