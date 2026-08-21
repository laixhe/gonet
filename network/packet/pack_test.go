package packet

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"testing"
)

func TestPackUnpack(t *testing.T) {
	msg := NewMessage(100, []byte("hello"))
	data, err := Pack(msg)
	if err != nil {
		t.Fatalf("Pack 失败: %v", err)
	}
	got, err := Unpack(data)
	if err != nil {
		t.Fatalf("Unpack 失败: %v", err)
	}
	if got.ID != 100 || got.DataLen != 5 || string(got.Data) != "hello" {
		t.Errorf("Unpack = %+v, want ID=100 DataLen=5 Data=hello", got)
	}
}

func TestPackUnpackEmpty(t *testing.T) {
	data, err := Pack(NewMessage(1, nil))
	if err != nil {
		t.Fatalf("Pack 失败: %v", err)
	}
	got, err := Unpack(data)
	if err != nil {
		t.Fatalf("Unpack 失败: %v", err)
	}
	if got.ID != 1 || got.DataLen != 0 || got.Data != nil {
		t.Errorf("Unpack = %+v, want 空消息", got)
	}
}

func TestUnpackTooLarge(t *testing.T) {
	var buf bytes.Buffer
	_ = binary.Write(&buf, byteOrder(), uint32(1))
	_ = binary.Write(&buf, byteOrder(), MaxMessageLen+1)
	if _, err := Unpack(buf.Bytes()); !errors.Is(err, ErrMessageTooLarge) {
		t.Errorf("err = %v, want ErrMessageTooLarge", err)
	}
}

func TestTcpBufReadTooLarge(t *testing.T) {
	var buf bytes.Buffer
	_ = binary.Write(&buf, byteOrder(), uint32(1))
	_ = binary.Write(&buf, byteOrder(), MaxMessageLen+1)
	if _, err := TcpBufRead(bufio.NewReader(&buf)); !errors.Is(err, ErrMessageTooLarge) {
		t.Errorf("err = %v, want ErrMessageTooLarge", err)
	}
}

// chunkReader 分块提供数据, 模拟 TCP 半包到达
type chunkReader struct {
	chunks [][]byte
	pos    int
}

func (r *chunkReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.chunks) {
		return 0, io.EOF
	}
	n := copy(p, r.chunks[r.pos])
	r.pos++
	return n, nil
}

func TestTcpBufReadStickyPacket(t *testing.T) {
	// 粘包: 多条消息在一个流中连续到达, 应依次解出
	want := make([]*Message, 3)
	var stream []byte
	for i := 0; i < 3; i++ {
		msg := NewMessage(uint32(i+1), []byte(fmt.Sprintf("msg-%d", i+1)))
		data, err := Pack(msg)
		if err != nil {
			t.Fatalf("Pack 失败: %v", err)
		}
		stream = append(stream, data...)
		want[i] = msg
	}

	reader := bufio.NewReader(bytes.NewReader(stream))
	for i, w := range want {
		got, err := TcpBufRead(reader)
		if err != nil {
			t.Fatalf("第 %d 条消息读取失败: %v", i, err)
		}
		if got.ID != w.ID || string(got.Data) != string(w.Data) {
			t.Errorf("第 %d 条 = ID:%d data:%s, want ID:%d data:%s", i, got.ID, got.Data, w.ID, w.Data)
		}
	}
	// 流耗尽后应返回错误
	if _, err := TcpBufRead(reader); err == nil {
		t.Error("流耗尽后应返回错误")
	}
}

func TestTcpBufReadAcrossBoundary(t *testing.T) {
	// 消息数据跨 bufio 缓冲边界(>4096 字节), 应正确完整解出
	msg := NewMessage(7, bytes.Repeat([]byte("x"), 5000))
	data, err := Pack(msg)
	if err != nil {
		t.Fatalf("Pack 失败: %v", err)
	}
	reader := bufio.NewReader(bytes.NewReader(data))
	got, err := TcpBufRead(reader)
	if err != nil {
		t.Fatalf("TcpBufRead 失败: %v", err)
	}
	if got.ID != 7 || got.DataLen != 5000 || len(got.Data) != 5000 {
		t.Errorf("解包结果不符: ID=%d DataLen=%d len=%d", got.ID, got.DataLen, len(got.Data))
	}
}

func TestTcpBufReadHalfPacket(t *testing.T) {
	// 半包: 数据分块到达且最终完整, 应正确解出
	msg := NewMessage(9, []byte("half-packet-test"))
	data, err := Pack(msg)
	if err != nil {
		t.Fatalf("Pack 失败: %v", err)
	}
	// 切成 3 块喂给 reader
	third := len(data) / 3
	reader := bufio.NewReader(&chunkReader{chunks: [][]byte{
		data[:third], data[third : 2*third], data[2*third:],
	}})
	got, err := TcpBufRead(reader)
	if err != nil {
		t.Fatalf("TcpBufRead 失败: %v", err)
	}
	if got.ID != 9 || string(got.Data) != "half-packet-test" {
		t.Errorf("解包结果不符: ID=%d data=%s", got.ID, got.Data)
	}
}

func TestTcpBufReadIncompleteEOF(t *testing.T) {
	// 数据不足且流已结束(EOF), 应返回错误而非 panic 或死循环
	msg := NewMessage(1, []byte("hello world"))
	data, err := Pack(msg)
	if err != nil {
		t.Fatalf("Pack 失败: %v", err)
	}
	half := len(data) / 2
	reader := bufio.NewReader(bytes.NewReader(data[:half]))
	if _, err := TcpBufRead(reader); err == nil {
		t.Fatal("半包数据应返回错误")
	}
}
