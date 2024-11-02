package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/laixhe/gonet/configx"
	"github.com/laixhe/gonet/logx"
	"github.com/laixhe/gonet/network/packet"
	"github.com/laixhe/gonet/proto/gen/config/clog"
)

var (
	// GitVersion 指定版本号 ( go build -ldflags "-X main.GitVersion=2024-11-01" )
	GitVersion string
)

var (
	// flagConfigFile 指定配置文件 (tcpclient --config=./config.yaml)
	flagConfigFile string
)

func main() {
	// init config
	flag.StringVar(&flagConfigFile, "config", "./config.yaml", "config path: --config config.yaml")
	flag.Parse()
	fmt.Println("main show", flagConfigFile, GitVersion)
	// init config
	config := struct {
		Log *clog.Log `mapstructure:"log"`
	}{}
	configx.Init(flagConfigFile, false, &config)
	logx.Init(config.Log)
	// 向服务端建立链接
	conn, err := net.Dial("tcp", "127.0.0.1:5050")
	if err != nil {
		log.Fatal("无法建立链接：", err)
	}
	defer conn.Close() //关闭链接

	logx.Debug("链接建立成功！")

	go func() {
		for {
			data, err := packet.TcpRead(conn)
			if err != nil {
				logx.Errorf("接收服务端数据失败： %s", err)
				return
			}
			logx.Infof("%d %d %s", data.DataLen, data.ID, string(data.Data))
		}
	}()

	go func() {
		for {
			// 向服务发送数据
			wData := "是的! " + fmt.Sprintf("%v", time.Now().UnixMilli())
			data, err := packet.Pack(packet.NewMessage(111, []byte(wData)))
			if err != nil {
				logx.Errorf("Pack data 失败： %s", err)
				return
			}
			_, err = conn.Write(data)
			if err != nil {
				logx.Errorf("向服务发送数据失败： %s", err)
				return
			}

			//log.Printf("向服务发送：%s (共 %d 字节)", wData, wLen)
			time.Sleep(time.Millisecond * 10)
		}
	}()

	select {}
}
