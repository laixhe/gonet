package main

import "github.com/laixhe/gonet/network/tcp"

func main() {
	if err := tcp.NewServer().Start(":5050"); err != nil {
		panic(err)
	}
}
