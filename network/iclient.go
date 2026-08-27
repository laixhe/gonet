package network

// IClient 客户端接口
type IClient interface {
	Start(addr string) error                         // 连接并启动
	Stop() error                                     // 停止
	Send(id uint32, data []byte) error               // 发送消息
	SetHandler(handler func(id uint32, data []byte)) // 设置消息处理
}
