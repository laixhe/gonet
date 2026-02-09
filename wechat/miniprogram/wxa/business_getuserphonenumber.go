package wxa

import (
	"errors"
	"fmt"

	"resty.dev/v3"
)

type GetUserPhoneNumberWatermark struct {
	Timestamp string `json:"timestamp"` // 用户获取手机号操作的时间戳
	Appid     string `json:"appid"`     // 小程序 appid
}

type GetUserPhoneNumberPhoneInfo struct {
	PhoneNumber     string                      `json:"phoneNumber"`     // 用户绑定的手机号（国外手机号会有区号）
	PurePhoneNumber string                      `json:"purePhoneNumber"` // 没有区号的手机号
	CountryCode     string                      `json:"countryCode"`     // 区号
	Watermark       GetUserPhoneNumberWatermark `json:"watermark"`       // 数据水印
}

type GetUserPhoneNumberResponse struct {
	ErrCode   int                         `json:"errcode"`    // 错误码(0 成功)(-1 系统繁忙)(40001 无效access_token)
	ErrMsg    string                      `json:"errmsg"`     // 错误信息，请求失败时返回
	PhoneInfo GetUserPhoneNumberPhoneInfo `json:"phone_info"` // 用户手机号信息
}

// GetUserPhoneNumber 获取手机号(getPhoneNumber)
// DOC https://developers.weixin.qq.com/miniprogram/dev/server/API/user-info/phone-number/api_getphonenumber.html
// POST https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=ACCESS_TOKEN
// BODY {"code":"XXX"}
func GetUserPhoneNumber(httpClient *resty.Client, accessToken, code string) (*GetUserPhoneNumberResponse, error) {
	httpResp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"access_token": accessToken,
		}).
		SetBody(map[string]string{
			"code": code,
		}).
		SetResult(&GetUserPhoneNumberResponse{}).
		SetForceResponseContentType("application/json").
		Post("/wxa/business/getuserphonenumber")
	if err != nil {
		return &GetUserPhoneNumberResponse{ErrCode: -1, ErrMsg: err.Error()}, err
	}
	if httpResp.IsSuccess() {
		resp, is := httpResp.Result().(*GetUserPhoneNumberResponse)
		if is {
			if resp.ErrCode != 0 {
				return resp, fmt.Errorf("%d %s", resp.ErrCode, resp.ErrMsg)
			}
			return resp, nil
		}
	}
	return &GetUserPhoneNumberResponse{
		ErrCode: httpResp.StatusCode(),
		ErrMsg:  httpResp.String(),
	}, errors.New(httpResp.String())
}
