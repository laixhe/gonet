module github.com/laixhe/gonet/sdk/wechat/work

go 1.26

replace github.com/laixhe/gonet/sdk/wechat/apiutil => ../apiutil

require (
	github.com/laixhe/gonet/sdk/wechat/apiutil v0.0.0-00010101000000-000000000000
	resty.dev/v3 v3.0.0-rc.3
)

require golang.org/x/net v0.43.0 // indirect
