// Package packet 提供二进制消息协议的定义与编解码, 消息格式为 [ID:4字节][DataLen:4字节][Data]。
package packet

// Message 消息
type Message struct {
	ID      uint32 // 消息的ID
	DataLen uint32 // 消息的长度
	Data    []byte // 消息的内容
}

// NewMessage 创建一个 Message 消息包
func NewMessage(ID uint32, data []byte) *Message {
	return &Message{
		ID:      ID,
		DataLen: uint32(len(data)),
		Data:    data,
	}
}

// Init 初始化消息字段
func (msg *Message) Init(ID uint32, data []byte) {
	msg.ID = ID
	msg.Data = data
	msg.DataLen = uint32(len(data))
}
