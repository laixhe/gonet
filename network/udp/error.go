package udp

import "errors"

// ErrClosed 连接已关闭
var ErrClosed = errors.New("udp connection closed")
