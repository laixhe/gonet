module github.com/laixhe/gonet/network

go 1.26

replace github.com/laixhe/gonet/xlog => ../xlog

require (
	github.com/gorilla/websocket v1.5.3
	github.com/laixhe/gonet/xlog v0.0.0-00010101000000-000000000000
	github.com/xtaci/kcp-go/v5 v5.6.72
)

require (
	github.com/klauspost/cpuid/v2 v2.2.6 // indirect
	github.com/klauspost/reedsolomon v1.12.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
