package oauth2

import (
	"errors"

	"resty.dev/v3"
)

type AccessTokenResponse struct {
	Errcode      int    `json:"errcode"`
	Errmsg       string `json:"errmsg"`
	AccessToken  string `json:"access_token"`  // 接口调用凭证
	ExpiresIn    int    `json:"expires_in"`    // access_token 接口调用凭证超时时间，单位（秒）
	RefreshToken string `json:"refresh_token"` // 用户刷新 access_token
	Openid       string `json:"openid"`        // 授权用户唯一标识
	Scope        string `json:"scope"`         // 用户授权的作用域，使用逗号（,）分隔
	Unionid      string `json:"unionid"`       // 当且仅当该网站应用已获得该用户的 userinfo 授权时，才会出现该字段
}

// AccessToken 通过 code 获取 access_token
// DOC WEB https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html
// DOC APP https://developers.weixin.qq.com/doc/oplatform/Mobile_App/WeChat_Login/Development_Guide.html
// GET https://api.weixin.qq.com/sns/oauth2/access_token?appid=APPID&secret=SECRET&code=CODE&grant_type=authorization_code
func AccessToken(httpClient *resty.Client, appid, secret, code string) (*AccessTokenResponse, error) {
	res, err := httpClient.R().
		SetQueryParams(map[string]string{
			"appid":      appid,
			"secret":     secret,
			"code":       code,
			"grant_type": "authorization_code",
		}).
		SetResult(&AccessTokenResponse{}).
		SetError(&AccessTokenResponse{}).
		Get("/sns/oauth2/access_token")
	if err != nil {
		return nil, err
	}
	if res.IsSuccess() {
		accessTokenResponse, is := res.Result().(*AccessTokenResponse)
		if is {
			if accessTokenResponse.Errcode != 0 {
				return nil, errors.New(accessTokenResponse.Errmsg)
			}
			return accessTokenResponse, nil
		}
	}
	if res.IsError() {
		accessTokenResponse, is := res.Error().(*AccessTokenResponse)
		if is {
			return nil, errors.New(accessTokenResponse.Errmsg)
		}
	}
	return nil, errors.New("获取 access_token 失败：" + res.String())
}
