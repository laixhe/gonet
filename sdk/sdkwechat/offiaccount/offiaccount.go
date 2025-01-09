package offiaccount

import (
	"errors"

	"github.com/ArtisanCloud/PowerSocialite/v3/src/providers"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"

	"github.com/laixhe/gonet/protocol/gen/config/cwechat"
	"github.com/laixhe/gonet/xlog"
)

// 微信公众号

type SdkWeChatOffiaccount struct {
	c      *cwechat.Offiaccount
	client *officialAccount.OfficialAccount
}

var sdkOffiaccount *SdkWeChatOffiaccount

// Init 初始化公众号
func Init(c *cwechat.Offiaccount) error {
	if c == nil {
		return errors.New("wechat offiaccount config as nil")
	}
	if c.AppId == "" {
		return errors.New("wechat offiaccount config appid as empty")
	}
	if c.Secret == "" {
		return errors.New("wechat offiaccount config secret as empty")
	}
	xlog.Debugf("wechat offiaccount config=%v", c)
	// doc https://powerwechat.artisan-cloud.com/zh/official-account
	client, err := officialAccount.NewOfficialAccount(&officialAccount.UserConfig{
		AppID:  c.AppId,
		Secret: c.Secret,
		OAuth: officialAccount.OAuth{
			Scopes: []string{"snsapi_userinfo"},
		},
		HttpDebug: true,
		Debug:     true,
		Log:       officialAccount.Log{Stdout: true},
	})
	if err != nil {
		return err
	}
	//
	sdkOffiaccount = &SdkWeChatOffiaccount{
		c:      c,
		client: client,
	}
	return nil
}

// AccessToken APP微信登录(通过 code 获取用户 access_token)
func AccessToken(code string) (*providers.User, error) {
	return sdkOffiaccount.client.OAuth.UserFromCode(code)
}

// UserInfo 获取用户基本信息
func UserInfo(openId, accessToken string) (*providers.User, error) {
	return sdkOffiaccount.client.OAuth.UserFromToken(accessToken, openId)
}
