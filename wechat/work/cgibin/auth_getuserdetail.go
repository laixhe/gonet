package cgibin

import (
	"errors"
	"fmt"

	"resty.dev/v3"
)

type GetUserDetailResponse struct {
	ErrCode int    `json:"errcode"`  // 出错返回码，(0 成功)(-1 系统繁忙)
	ErrMsg  string `json:"errmsg"`   // 返回码提示语
	Userid  string `json:"userid"`   // 成员 UserID
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
func GetUserDetail(httpClient *resty.Client, accessToken, userTicket string) (*GetUserDetailResponse, error) {
	httpResp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"access_token": accessToken,
		}).
		SetBody(map[string]string{
			"user_ticket": userTicket,
		}).
		SetResult(&GetUserDetailResponse{}).
		SetForceResponseContentType("application/json").
		Post("/cgi-bin/auth/getuserdetail")
	if err != nil {
		return &GetUserDetailResponse{ErrCode: -1, ErrMsg: err.Error()}, err
	}
	if httpResp.IsSuccess() {
		resp, is := httpResp.Result().(*GetUserDetailResponse)
		if is {
			if resp.ErrCode != 0 {
				return resp, fmt.Errorf("%d %s", resp.ErrCode, resp.ErrMsg)
			}
			return resp, nil
		}
	}
	return &GetUserDetailResponse{
		ErrCode: httpResp.StatusCode(),
		ErrMsg:  httpResp.String(),
	}, errors.New(httpResp.String())
}
