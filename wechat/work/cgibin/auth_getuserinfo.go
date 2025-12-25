package cgibin

import (
	"errors"
	"fmt"

	"resty.dev/v3"
)

type GetUserInfoResponse struct {
	ErrCode        int    `json:"errcode"`         // 出错返回码，(0 成功)(-1 系统繁忙)(40029 invalid code)
	ErrMsg         string `json:"errmsg"`          // 返回码提示语
	Userid         string `json:"userid"`          // 成员 UserID
	UserTicket     string `json:"user_ticket"`     // 成员票据，最大为 512 字节，有效期为 1800s
	Openid         string `json:"openid"`          // 非企业成员的标识，对当前企业唯一
	ExternalUserid string `json:"external_userid"` // 外部联系人 id
}

// GetUserInfo 获取访问用户身份
// DOC https://developer.work.weixin.qq.com/document/path/91023
// GET https://qyapi.weixin.qq.com/cgi-bin/auth/getuserinfo?access_token=ACCESS_TOKEN&code=CODE
func GetUserInfo(httpClient *resty.Client, accessToken, code string) (*GetUserInfoResponse, error) {
	httpResp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"access_token": accessToken,
			"code":         code,
		}).
		SetResult(&GetUserInfoResponse{}).
		SetForceResponseContentType("application/json").
		Get("/cgi-bin/auth/getuserinfo")
	if err != nil {
		return &GetUserInfoResponse{ErrCode: -1, ErrMsg: err.Error()}, err
	}
	if httpResp.IsSuccess() {
		resp, is := httpResp.Result().(*GetUserInfoResponse)
		if is {
			if resp.ErrCode != 0 {
				return resp, fmt.Errorf("%d %s", resp.ErrCode, resp.ErrMsg)
			}
			return resp, nil
		}
	}
	return &GetUserInfoResponse{
		ErrCode: httpResp.StatusCode(),
		ErrMsg:  httpResp.String(),
	}, errors.New(httpResp.String())
}
