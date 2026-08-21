package cgibin

import (
	"context"

	"resty.dev/v3"

	"github.com/laixhe/gonet/sdk/wechat/work/internal/apiutil"
)

type GetUserInfoResponse struct {
	ErrCode        int    `json:"errcode"`         // 出错返回码，(0 成功)(-1 系统繁忙)(40029 invalid code)
	ErrMsg         string `json:"errmsg"`          // 返回码提示语
	UserID         string `json:"userid"`          // 成员 UserID
	UserTicket     string `json:"user_ticket"`     // 成员票据，最大为 512 字节，有效期为 1800s
	OpenID         string `json:"openid"`          // 非企业成员的标识，对当前企业唯一
	ExternalUserID string `json:"external_userid"` // 外部联系人 id
}

// GetUserInfo 获取访问用户身份
// DOC https://developer.work.weixin.qq.com/document/path/91023
// GET https://qyapi.weixin.qq.com/cgi-bin/auth/getuserinfo?access_token=ACCESS_TOKEN&code=CODE
func GetUserInfo(ctx context.Context, httpClient *resty.Client, accessToken, code string) (*GetUserInfoResponse, error) {
	return apiutil.Get[GetUserInfoResponse](ctx, httpClient, "/cgi-bin/auth/getuserinfo", map[string]string{
		"access_token": accessToken,
		"code":         code,
	})
}
