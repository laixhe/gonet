package tcp

import (
	"net"
	"time"

	"github.com/laixhe/gonet/network"
)

var (
	// DefaultHeartbeatInterval 默认心跳发送/检测间隔
	DefaultHeartbeatInterval = 15 * time.Second
	// DefaultHeartbeatTimeout 默认心跳超时时间, 超过该时间未收到任何消息则断开连接
	DefaultHeartbeatTimeout = 30 * time.Second
)

// Connection 用户链接, 复用 network.StreamConnection 通用流式连接
type Connection struct {
	*network.StreamConnection
}

var _ network.IConn = &Connection{}

// init 初始化连接字段
func (c *Connection) init(conn net.Conn, manager network.IManager, id int64, dispatch func(network.IConn, uint32, []byte), heartbeatInterval, heartbeatTimeout, writeTimeout time.Duration, workerCount int) {
	c.StreamConnection = network.NewStreamConnection(conn, manager, id, dispatch, heartbeatInterval, heartbeatTimeout, writeTimeout, workerCount, "tcp")
}
