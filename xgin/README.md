```
package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/laixhe/gonet/xgin"
	"github.com/laixhe/gonet/xlog"
)

func main() {
	logs, err := xlog.Init(&xlog.Config{
		CallerSkip: 1,
	})
	if err != nil {
		panic(err)
	}
	app := xgin.New(true, logs.Logger()).
		UseCors().
		UseRecovery()

	t := app.App().Group("test")
	t.POST("123", func(ctx *gin.Context) {
		maps := make(map[string]interface{})
		ctx.ShouldBindJSON(&maps)
		logs.Debug("T---xxxx---", xgin.ZapField(ctx)...)
		ctx.String(200, "gin test/123")
	})

	fmt.Println(app.Listen(":8010"))
}
```