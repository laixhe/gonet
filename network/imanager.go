package network

// IManager 连接管理器接口
type IManager interface {
	Add(conn IConn) error        // 添加链接
	Remove(conn IConn)           // 删除连接
	Close()                      // 关闭连接
	Count() int64                // 当前连接数
	FindByID(id int64) IConn     // 按 ID 查找连接, 未找到返回 nil
	FindByUid(uid int64) IConn   // 按 Uid 查找连接, 未找到返回 nil
	KickByID(id int64)           // 按 ID 踢下线(关闭连接)
	KickByUid(uid int64)         // 按 Uid 踢下线(关闭连接)
	ForEach(fn func(conn IConn)) // 遍历所有连接
}
