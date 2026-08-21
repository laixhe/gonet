package sns

import (
	"context"

	"resty.dev/v3"

	"github.com/laixhe/gonet/sdk/wechat/miniprogram/internal/apiutil"
)

type JsCode2SessionResponse struct {
	ErrCode    int    `json:"errcode"`     // 错误码(0 成功)(-1 系统繁忙)(40029 code无效)(45011 调用太频繁)(40226 高风险等级用户)
	ErrMsg     string `json:"errmsg"`      // 错误信息，请求失败时返回
	SessionKey string `json:"session_key"` // 会话密钥
	UnionID    string `json:"unionid"`     // 开放平台的唯一标识符
	OpenID     string `json:"openid"`      // 授权用户唯一标识
}

// JsCode2Session 小程序登录
// 通过 code 获取 openid
// DOC https://developers.weixin.qq.com/miniprogram/dev/server/API/user-login/api_code2session.html
// GET https://api.weixin.qq.com/sns/jscode2session?appid=APPID&secret=SECRET&js_code=JS_CODE&grant_type=GRANT_TYPE
func JsCode2Session(ctx context.Context, httpClient *resty.Client, appID, secret, code string) (*JsCode2SessionResponse, error) {
	return apiutil.Get[JsCode2SessionResponse](ctx, httpClient, "/sns/jscode2session", map[string]string{
		"appid":      appID,
		"secret":     secret,
		"js_code":    code,
		"grant_type": "authorization_code",
	})
}
