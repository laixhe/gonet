package network

import "net"

// IManager 连接管理器接口
type IManager interface {
	Add(net.Conn) error // 添加链接
	Remove(net.Conn)    // 删除连接
	Close()             // 关闭连接
}
