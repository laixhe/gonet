package oauth2

import (
	"errors"
	"fmt"

	"resty.dev/v3"
)

type AccessTokenResponse struct {
	Errcode      int    `json:"errcode"`
	Errmsg       string `json:"errmsg"`
	AccessToken  string `json:"access_token"`  // 接口调用凭证
	Unionid      string `json:"unionid"`       // 开放平台的唯一标识符
	Openid       string `json:"openid"`        // 授权用户唯一标识
	RefreshToken string `json:"refresh_token"` // 刷新 access_token 凭证
	ExpiresIn    int    `json:"expires_in"`    // 凭证有效时间，单位：秒。目前是7200秒之内的值(2个小时)
	Scope        string `json:"scope"`         // 用户授权的作用域，使用逗号（,）分隔
}

// AccessToken 微信登录
// 通过 code 获取 access_token
// DOC WEB https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html
// DOC APP https://developers.weixin.qq.com/doc/oplatform/Mobile_App/WeChat_Login/Development_Guide.html
// GET https://api.weixin.qq.com/sns/oauth2/access_token?appid=APPID&secret=SECRET&code=CODE&grant_type=authorization_code
func AccessToken(httpClient *resty.Client, appid, secret, code string) (*AccessTokenResponse, error) {
	httpResp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"appid":      appid,
			"secret":     secret,
			"code":       code,
			"grant_type": "authorization_code",
		}).
		SetResult(&AccessTokenResponse{}).
		SetForceResponseContentType("application/json").
		Get("/sns/oauth2/access_token")
	if err != nil {
		return &AccessTokenResponse{
			Errcode: httpResp.StatusCode(),
			Errmsg:  err.Error(),
		}, err
	}
	if httpResp.IsSuccess() {
		resp, is := httpResp.Result().(*AccessTokenResponse)
		if is {
			if resp.Errcode != 0 {
				return resp, fmt.Errorf("%d %s", resp.Errcode, resp.Errmsg)
			}
			return resp, nil
		}
	}
	return &AccessTokenResponse{
		Errcode: httpResp.StatusCode(),
		Errmsg:  httpResp.String(),
	}, errors.New(httpResp.String())
}
