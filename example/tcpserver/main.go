package main

import (
	"flag"
	"fmt"

	"github.com/laixhe/gonet/network/tcp"
	"github.com/laixhe/gonet/protocol/gen/config/clog"
	"github.com/laixhe/gonet/xconfig"
	"github.com/laixhe/gonet/xlog"
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
	xconfig.Init(flagConfigFile, false, &config)
	xlog.Init(config.Log)
	// server
	if err := tcp.NewServer().Start(":5050"); err != nil {
		panic(err)
	}
}
