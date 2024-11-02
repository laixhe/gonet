package main

import (
	"flag"
	"fmt"

	"github.com/laixhe/gonet/configx"
	"github.com/laixhe/gonet/logx"
	"github.com/laixhe/gonet/network/tcp"
	"github.com/laixhe/gonet/proto/gen/config/clog"
)

var (
	// GitVersion 指定版本号 ( go build -ldflags "-X main.GitVersion=2024-11-01" )
	GitVersion string
)

var (
	// flagConfigFile 指定配置文件 (tcpserver --config=./config.yaml)
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
	// server
	if err := tcp.NewServer().Start(":5050"); err != nil {
		panic(err)
	}
}
