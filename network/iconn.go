package network

const (
	ConnOpened int32 = 1 // 连接打开
	ConnHanged int32 = 2 // 连接挂起
	ConnClosed int32 = 3 // 连接关闭

	// MessageIDHeartbeat 心跳消息ID, 客户端自动发送, 服务端用于刷新心跳时间
	MessageIDHeartbeat uint32 = 1
)

// IConn 连接接口
type IConn interface {
	ID() int64                         // 获取当前连接ID
	Uid() int64                        // 获取用户ID
	BindUid(int64)                     // 绑定用户ID
	UnbindUid()                        // 解绑用户ID
	State() int32                      // 获取连接状态
	IsClosed() bool                    // 是否连接关闭
	RemoteAddr() string                // 获取远程地址(ip:port)
	Stop() error                       // 停止连接，结束当前连接
	Send(id uint32, data []byte) error // 发送消息
}
