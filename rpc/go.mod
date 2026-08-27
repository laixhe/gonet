module github.com/laixhe/gonet/rpc

go 1.26

// 排除被 workspace 内 tjfoc/gmsm(→旧版 grpc) 引入的老整包 genproto，
// 其包路径已由 google.golang.org/genproto/googleapis/{api,rpc} 提供，避免 ambiguous import。
exclude google.golang.org/genproto v0.0.0-20180817151627-c66870c02cf8

exclude google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55

require (
	go.etcd.io/etcd/api/v3 v3.7.1
	go.etcd.io/etcd/client/v3 v3.7.1
	google.golang.org/grpc v1.83.2
)

require (
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.7.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.7.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
