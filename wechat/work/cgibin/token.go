package cgibin

import (
	"errors"
	"fmt"

	"resty.dev/v3"
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
func GetToken(httpClient *resty.Client, corpid, corpsecret string) (*GetTokenResponse, error) {
	httpResp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"corpid":     corpid,
			"corpsecret": corpsecret,
		}).
		SetResult(&GetTokenResponse{}).
		SetForceResponseContentType("application/json").
		Get("/cgi-bin/gettoken")
	if err != nil {
		return &GetTokenResponse{ErrCode: -1, ErrMsg: err.Error()}, err
	}
	if httpResp.IsSuccess() {
		resp, is := httpResp.Result().(*GetTokenResponse)
		if is {
			if resp.ErrCode != 0 {
				return resp, fmt.Errorf("%d %s", resp.ErrCode, resp.ErrMsg)
			}
			return resp, nil
		}
	}
	return &GetTokenResponse{
		ErrCode: httpResp.StatusCode(),
		ErrMsg:  httpResp.String(),
	}, errors.New(httpResp.String())
}
