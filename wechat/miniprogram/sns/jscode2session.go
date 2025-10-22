package sns

import (
	"errors"
	"fmt"

	"resty.dev/v3"
)

type Jscode2sessionResponse struct {
	Errcode    int    `json:"errcode"`     // 错误信息，请求失败时返回(-1 系统繁忙)(40029 code无效)(45011 调用太频繁)(40226 高风险等级用户)
	Errmsg     string `json:"errmsg"`      // 错误码，请求失败时返回
	SessionKey string `json:"session_key"` // 会话密钥
	Unionid    string `json:"unionid"`     // 开放平台的唯一标识符
	Openid     string `json:"openid"`      // 授权用户唯一标识
}

// Jscode2session 小程序登录
// 通过 code 获取 openid
// DOC https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-login/code2Session.html
// GET https://api.weixin.qq.com/sns/jscode2session?appid=APPID&secret=SECRET&js_code=JSCODE&grant_type=authorization_code
func Jscode2session(httpClient *resty.Client, appid, secret, code string) (*Jscode2sessionResponse, error) {
	httpResp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"appid":      appid,
			"secret":     secret,
			"js_code":    code,
			"grant_type": "authorization_code",
		}).
		SetResult(&Jscode2sessionResponse{}).
		SetForceResponseContentType("application/json").
		Get("/sns/jscode2session")
	if err != nil {
		return &Jscode2sessionResponse{
			Errcode: httpResp.StatusCode(),
			Errmsg:  err.Error(),
		}, err
	}
	if httpResp.IsSuccess() {
		resp, is := httpResp.Result().(*Jscode2sessionResponse)
		if is {
			if resp.Errcode != 0 {
				return resp, fmt.Errorf("%d %s", resp.Errcode, resp.Errmsg)
			}
			return resp, nil
		}
	}
	return &Jscode2sessionResponse{
		Errcode: httpResp.StatusCode(),
		Errmsg:  httpResp.String(),
	}, errors.New(httpResp.String())
}
