package cgibin

import (
	"context"

	"resty.dev/v3"

	"github.com/laixhe/gonet/sdk/wechat/work/internal/apiutil"
)

type GetUserDetailResponse struct {
	ErrCode int    `json:"errcode"`  // 出错返回码，(0 成功)(-1 系统繁忙)
	ErrMsg  string `json:"errmsg"`   // 返回码提示语
	UserID  string `json:"userid"`   // 成员 UserID
	Gender  string `json:"gender"`   // 性别 0表示未定义 1表示男性 2表示女性 仅在用户同意 snsapi_privateinfo 授权时返回真实值，否则返回0
	Avatar  string `json:"avatar"`   // 头像 url 仅在用户同意 snsapi_privateinfo 授权时返回真实头像，否则返回默认头像
	QrCode  string `json:"qr_code"`  // 员工个人二维码（扫描可添加为外部联系人），仅在用户同意 snsapi_privateinfo 授权时返回
	Mobile  string `json:"mobile"`   // 手机，仅在用户同意 snsapi_privateinfo 授权时返回，第三方应用不可获
	Email   string `json:"email"`    // 邮箱，仅在用户同意 snsapi_privateinfo 授权时返回，第三方应用不可获取
	BizMail string `json:"biz_mail"` // 企业邮箱，仅在用户同意 snsapi_privateinfo 授权时返回，第三方应用不可获取
	Address string `json:"address"`  // 仅在用户同意 snsapi_privateinfo 授权时返回，第三方应用不可获取
}

// GetUserDetail 获取访问用户敏感信息
// DOC  https://developer.work.weixin.qq.com/document/path/95833
// POST https://qyapi.weixin.qq.com/cgi-bin/auth/getuserdetail?access_token=ACCESS_TOKEN
// BODY {"user_ticket":"XXX"}
func GetUserDetail(ctx context.Context, httpClient *resty.Client, accessToken, userTicket string) (*GetUserDetailResponse, error) {
	return apiutil.Post[GetUserDetailResponse](ctx, httpClient, "/cgi-bin/auth/getuserdetail", map[string]string{
		"access_token": accessToken,
	}, map[string]string{
		"user_ticket": userTicket,
	})
}
