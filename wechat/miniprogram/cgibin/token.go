package cgibin

import (
	"errors"
	"fmt"

	"resty.dev/v3"
)

type TokenResponse struct {
	Errcode     int    `json:"errcode"`      // 错误信息，请求失败时返回(-1 系统繁忙)(40164 IP白名单)(50004 禁止使用)(50007 账号已冻结)
	Errmsg      string `json:"errmsg"`       // 错误码，请求失败时返回
	AccessToken string `json:"access_token"` // 获取到的凭证
	ExpiresIn   int64  `json:"expires_in"`   // 凭证有效时间，单位：秒。目前是7200秒之内的值(2个小时)
}

// Token 获取接口调用凭据(getAccessToken)
// DOC https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/mp-access-token/getAccessToken.html
// GET https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=APPID&secret=APPSECRET
func Token(httpClient *resty.Client, appid, secret string) (*TokenResponse, error) {
	httpResp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"appid":      appid,
			"secret":     secret,
			"grant_type": "client_credential",
		}).
		SetResult(&TokenResponse{}).
		SetForceResponseContentType("application/json").
		Get("/cgi-bin/token")
	if err != nil {
		return &TokenResponse{
			Errcode: httpResp.StatusCode(),
			Errmsg:  err.Error(),
		}, err
	}
	if httpResp.IsSuccess() {
		resp, is := httpResp.Result().(*TokenResponse)
		if is {
			if resp.Errcode != 0 {
				return resp, fmt.Errorf("%d %s", resp.Errcode, resp.Errmsg)
			}
			return resp, nil
		}
	}
	return &TokenResponse{
		Errcode: httpResp.StatusCode(),
		Errmsg:  httpResp.String(),
	}, errors.New(httpResp.String())
}
