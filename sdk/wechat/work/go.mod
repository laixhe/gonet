module github.com/laixhe/gonet/sdk/wechat/work

go 1.26

require (
	github.com/laixhe/gonet/sdk/wechat/apiutil v0.0.0
	resty.dev/v3 v3.0.0-rc.3
)

require golang.org/x/net v0.43.0 // indirect

replace github.com/laixhe/gonet/sdk/wechat/apiutil => ../apiutil
