package network

// RouterHandler 消息处理函数
type RouterHandler func(conn IConn, data []byte)

// IServer 服务器接口
type IServer interface {
	Router(id uint32, handler RouterHandler)                            // 注册消息路由
	SetDefaultHandler(handler func(conn IConn, id uint32, data []byte)) // 设置默认消息处理, 未注册的消息走该处理器
	OnConnect(handler func(conn IConn))                                 // 设置连接建立回调
	OnDisconnect(handler func(conn IConn))                              // 设置连接断开回调
	Start(addr string) error                                            // 启动服务器
	Stop() error                                                        // 关闭服务器
	Addr() string                                                       // 获取实际监听地址, Start 前返回空串
	GetManager() IManager                                               // 获取连接管理器
	Broadcast(id uint32, data []byte)                                   // 广播消息给所有连接
	BroadcastExclude(id uint32, data []byte, excludeUid int64)          // 广播消息给所有连接, 排除指定 Uid
}
