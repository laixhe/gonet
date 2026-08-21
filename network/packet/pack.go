package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	DefaultHeaderLen uint32 = 8     // 包头长度
	ByteOrder        string = "big" // 消息字节序(大端/小端)
)

// MaxMessageLen 单条消息最大长度, 防止超大消息耗尽内存
var MaxMessageLen uint32 = 1 << 20 // 1MB

// ErrMessageTooLarge 消息超过最大长度限制
var ErrMessageTooLarge = errors.New("message too large")

// Pack 打包消息
func Pack(msg *Message) (packet []byte, err error) {
	var buf bytes.Buffer
	buf.Grow(len(msg.Data) + int(DefaultHeaderLen))

	if err = binary.Write(&buf, byteOrder(), msg.ID); err != nil {
		return
	}
	if err = binary.Write(&buf, byteOrder(), msg.DataLen); err != nil {
		return
	}
	if err = binary.Write(&buf, byteOrder(), msg.Data); err != nil {
		return
	}
	packet = buf.Bytes()
	return
}

// Unpack 解包消息
func Unpack(packet []byte) (msg *Message, err error) {
	msg = &Message{}
	var buf = bytes.NewBuffer(packet)

	if err = binary.Read(buf, byteOrder(), &msg.ID); err != nil {
		return
	}
	if err = binary.Read(buf, byteOrder(), &msg.DataLen); err != nil {
		return
	}
	if msg.DataLen > MaxMessageLen {
		err = ErrMessageTooLarge
		return
	}
	if msg.DataLen > 0 {
		msg.Data = make([]byte, msg.DataLen)
		if err = binary.Read(buf, byteOrder(), &msg.Data); err != nil {
			return
		}
	}
	return
}

func byteOrder() binary.ByteOrder {
	switch ByteOrder {
	case "big":
		return binary.BigEndian
	default:
		return binary.LittleEndian
	}
}
