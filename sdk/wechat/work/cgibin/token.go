package cgibin

import (
	"context"

	"resty.dev/v3"

	"github.com/laixhe/gonet/sdk/wechat/apiutil"
)

type GetTokenResponse struct {
	ErrCode     int    `json:"errcode"`      // 出错返回码，(0 成功)(-1 系统繁忙)
	ErrMsg      string `json:"errmsg"`       // 返回码提示语
	AccessToken string `json:"access_token"` // 获取到的凭证
	ExpiresIn   int64  `json:"expires_in"`   // 凭证有效时间，单位：秒。目前是7200秒之内的值(2个小时)
}

// GetToken 获取access_token
// DOC https://developer.work.weixin.qq.com/document/path/91039
// GET https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=ID&corpsecret=SECRET
func GetToken(ctx context.Context, httpClient *resty.Client, corpID, corpSecret string) (*GetTokenResponse, error) {
	return apiutil.Get[GetTokenResponse](ctx, httpClient, "/cgi-bin/gettoken", map[string]string{
		"corpid":     corpID,
		"corpsecret": corpSecret,
	})
}
