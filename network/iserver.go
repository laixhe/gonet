package network

// IServer 服务器接口
type IServer interface {
	RouterPath(path string)  // 路由路径
	Start(addr string) error // 启动服务器
	Stop() error             // 关闭服务器
	GetManager() IManager    // 获取连接管理器
}
