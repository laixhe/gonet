#### 例子
```
package main

import (
	"fmt"

	contribJwt "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/laixhe/gonet/xfiber"
	"github.com/laixhe/gonet/xlog"
)

func main() {
	logs, err := xlog.Init(&xlog.Config{
		CallerSkip: 1,
	})
	if err != nil {
		panic(err)
	}
	app := xfiber.New(logs.Logger()).
		UseCors().
		UseRecover()

	t := app.App().Group("test", xfiber.UseJwt(contribJwt.Config{
		SigningKey: contribJwt.SigningKey{Key: []byte("12345678")},
	}))

	t.Post("123", func(c fiber.Ctx) error {
		log.WithContext(c.Context()).Debug("T---xxxx---", string(c.Body()))
		return c.SendString("test/123")
	})

	fmt.Println(app.Listen(":8010"))
}
```