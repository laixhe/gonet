package cgibin

import (
	"context"

	"resty.dev/v3"

	"github.com/laixhe/gonet/sdk/wechat/miniprogram/internal/apiutil"
)

type TokenResponse struct {
	ErrCode     int    `json:"errcode"`      // 错误码(0 成功)(-1 系统繁忙)(40164 IP白名单)(50004 禁止使用)(50007 账号已冻结)
	ErrMsg      string `json:"errmsg"`       // 错误信息，请求失败时返回
	AccessToken string `json:"access_token"` // 获取到的凭证
	ExpiresIn   int64  `json:"expires_in"`   // 凭证有效时间，单位：秒。目前是7200秒之内的值(2个小时)
}

// Token 获取接口调用凭据(getAccessToken)
// DOC https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-access-token/api_getaccesstoken.html
// GET https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=APPID&secret=APPSECRET
func Token(ctx context.Context, httpClient *resty.Client, appID, secret string) (*TokenResponse, error) {
	return apiutil.Get[TokenResponse](ctx, httpClient, "/cgi-bin/token", map[string]string{
		"appid":      appID,
		"secret":     secret,
		"grant_type": "client_credential",
	})
}

// StableToken 获取稳定版接口调用凭据
// DOC https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-access-token/api_getstableaccesstoken.html
// POST https://api.weixin.qq.com/cgi-bin/stable_token
// BODY {"grant_type":"client_credential","appid":"APPID","secret":"APPSECRET","force_refresh":false}
func StableToken(ctx context.Context, httpClient *resty.Client, appID string, secret string, forceRefresh bool) (*TokenResponse, error) {
	return apiutil.Post[TokenResponse](ctx, httpClient, "/cgi-bin/stable_token", nil, map[string]interface{}{
		"appid":         appID,
		"secret":        secret,
		"force_refresh": forceRefresh,
		"grant_type":    "client_credential",
	})
}
