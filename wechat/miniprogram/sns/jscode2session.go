package sns

import (
	"errors"
	"fmt"

	"resty.dev/v3"
)

type JsCode2SessionResponse struct {
	ErrCode    int    `json:"errcode"`     // 错误码(0 成功)(-1 系统繁忙)(40029 code无效)(45011 调用太频繁)(40226 高风险等级用户)
	ErrMsg     string `json:"errmsg"`      // 错误信息，请求失败时返回
	SessionKey string `json:"session_key"` // 会话密钥
	Unionid    string `json:"unionid"`     // 开放平台的唯一标识符
	Openid     string `json:"openid"`      // 授权用户唯一标识
}

// JsCode2Session 小程序登录
// 通过 code 获取 openid
// DOC https://developers.weixin.qq.com/miniprogram/dev/server/API/user-login/api_code2session.html
// GET https://api.weixin.qq.com/sns/jscode2session?appid=APPID&secret=SECRET&js_code=JS_CODE&grant_type=GRANT_TYPE
func JsCode2Session(httpClient *resty.Client, appid, secret, code string) (*JsCode2SessionResponse, error) {
	httpResp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"appid":      appid,
			"secret":     secret,
			"js_code":    code,
			"grant_type": "authorization_code",
		}).
		SetResult(&JsCode2SessionResponse{}).
		SetForceResponseContentType("application/json").
		Get("/sns/jscode2session")
	if err != nil {
		return &JsCode2SessionResponse{ErrCode: -1, ErrMsg: err.Error()}, err
	}
	if httpResp.IsSuccess() {
		resp, is := httpResp.Result().(*JsCode2SessionResponse)
		if is {
			if resp.ErrCode != 0 {
				return resp, fmt.Errorf("%d %s", resp.ErrCode, resp.ErrMsg)
			}
			return resp, nil
		}
	}
	return &JsCode2SessionResponse{
		ErrCode: httpResp.StatusCode(),
		ErrMsg:  httpResp.String(),
	}, errors.New(httpResp.String())
}
