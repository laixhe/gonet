package kcp

import (
	"net"
	"time"

	"github.com/laixhe/gonet/network"
	kcpv5 "github.com/xtaci/kcp-go/v5"
)

// Connection 用户链接, 复用 network.StreamConnection 通用流式连接
type Connection struct {
	*network.StreamConnection
}

var _ network.IConn = &Connection{}

// init 初始化连接字段, 并对 KCP 会话设置快速模式与 MTU
func (c *Connection) init(conn net.Conn, manager network.IManager, id int64, dispatch func(network.IConn, uint32, []byte), noDelay bool, mtu int, heartbeatInterval, heartbeatTimeout, writeTimeout time.Duration, workerCount int) {
	c.StreamConnection = network.NewStreamConnection(conn, manager, id, dispatch, heartbeatInterval, heartbeatTimeout, writeTimeout, workerCount, "kcp")
	if sess, ok := conn.(*kcpv5.UDPSession); ok {
		if noDelay {
			sess.SetNoDelay(1, 10, 2, 1)
		}
		if mtu > 0 {
			sess.SetMtu(mtu)
		}
	}
}
