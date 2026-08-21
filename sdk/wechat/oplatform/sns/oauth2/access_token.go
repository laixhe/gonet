package oauth2

import (
	"context"

	"resty.dev/v3"

	"github.com/laixhe/gonet/sdk/wechat/apiutil"
)

type AccessTokenResponse struct {
	ErrCode      int    `json:"errcode"`       // 错误码(0 成功)(-1 系统繁忙)
	ErrMsg       string `json:"errmsg"`        // 错误信息，请求失败时返回
	AccessToken  string `json:"access_token"`  // 接口调用凭证
	UnionID      string `json:"unionid"`       // 开放平台的唯一标识符
	OpenID       string `json:"openid"`        // 授权用户唯一标识
	RefreshToken string `json:"refresh_token"` // 刷新 access_token 凭证
	ExpiresIn    int64  `json:"expires_in"`    // 凭证有效时间，单位：秒。目前是7200秒之内的值(2个小时)
	Scope        string `json:"scope"`         // 用户授权的作用域，使用逗号（,）分隔
}

// AccessToken 微信登录
// 通过 code 获取 access_token
// DOC WEB https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html
// DOC APP https://developers.weixin.qq.com/doc/oplatform/Mobile_App/WeChat_Login/Development_Guide.html
// GET https://api.weixin.qq.com/sns/oauth2/access_token?appid=APPID&secret=SECRET&code=CODE&grant_type=authorization_code
func AccessToken(ctx context.Context, httpClient *resty.Client, appID, secret, code string) (*AccessTokenResponse, error) {
	return apiutil.Get[AccessTokenResponse](ctx, httpClient, "/sns/oauth2/access_token", map[string]string{
		"appid":      appID,
		"secret":     secret,
		"code":       code,
		"grant_type": "authorization_code",
	})
}
