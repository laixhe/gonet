package douyin

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/alibabacloud-go/tea/tea"
	openApiSdkClient "github.com/bytedance/douyin-openapi-sdk-go/client"
)

type JsCode2SessionResponse struct {
	OpenID          string `json:"openid"`
	UnionID         string `json:"unionid"`
	SessionKey      string `json:"sessionKey"`
	AnonymousOpenID string `json:"anonymous_openid"`
}

// MiniLoginJsCode2Session 小程序登录
// 通常用于小程序通过临时登录凭证 code 换取用户唯一标识 openid unionid session_key 等信息
// DOC https://developer.open-douyin.com/docs/resource/zh-CN/mini-app/develop/server/basic-abilities/log-in/code-2-session
func (d *Douyin) MiniLoginJsCode2Session(code string, anonymousCode ...string) (*JsCode2SessionResponse, *ErrorData) {
	req := &openApiSdkClient.V2Jscode2sessionRequest{}
	req.SetAppid(d.config.AppID)
	req.SetSecret(d.config.AppSecret)
	if code != "" {
		req.SetCode(code)
	}
	if code == "" && len(anonymousCode) > 0 && len(anonymousCode[0]) > 0 {
		req.SetAnonymousCode(anonymousCode[0])
	}
	resp, err := d.client.V2Jscode2session(req)
	if err != nil {
		var sdkError *tea.SDKError
		switch {
		case errors.As(err, &sdkError):
			errorCode, _ := strconv.Atoi(tea.StringValue(sdkError.Code))
			return nil, NewErrorData(errorCode, tea.StringValue(sdkError.Message))
		default:
			return nil, NewErrorData(ECodeCall, err.Error())
		}
	}
	if code != "" {
		if resp.Data == nil || resp.Data.Openid == nil || resp.Data.Unionid == nil || *resp.Data.Openid == "" || *resp.Data.Unionid == "" {
			if resp.ErrNo == nil {
				resp.ErrNo = tea.Int64(ECodeCall)
			}
			if resp.ErrTips == nil {
				resp.ErrTips = tea.String("调用失败")
			}
			return nil, NewErrorData(int(tea.Int64Value(resp.ErrNo)), tea.StringValue(resp.ErrTips))
		}
	}
	if code == "" && len(anonymousCode) > 0 && len(anonymousCode[0]) > 0 {
		if resp.Data == nil || resp.Data.AnonymousOpenid == nil || *resp.Data.AnonymousOpenid == "" {
			if resp.ErrNo == nil {
				resp.ErrNo = tea.Int64(ECodeCall)
			}
			if resp.ErrTips == nil {
				resp.ErrTips = tea.String("调用失败")
			}
			return nil, NewErrorData(int(tea.Int64Value(resp.ErrNo)), tea.StringValue(resp.ErrTips))
		}
	}
	return &JsCode2SessionResponse{
		OpenID:          tea.StringValue(resp.Data.Openid),
		UnionID:         tea.StringValue(resp.Data.Unionid),
		SessionKey:      tea.StringValue(resp.Data.SessionKey),
		AnonymousOpenID: tea.StringValue(resp.Data.AnonymousOpenid),
	}, nil
}

type GetPhoneNumberResponse struct {
	PhoneNumber     string `json:"phoneNumber"`     // 用户绑定的手机号（国外手机号会有区号）
	PurePhoneNumber string `json:"purePhoneNumber"` // 没有区号的手机号
	CountryCode     string `json:"countryCode"`     // 区号
	Watermark       struct {
		AppID     string `json:"appid"`
		Timestamp int    `json:"timestamp"`
	} `json:"watermark"`
}

// MiniLoginGetPhoneNumberInfo 获取手机号
// 每个 code 只能使用一次，code 的有效期为 5 min
// DOC https://developer.open-douyin.com/docs/resource/zh-CN/mini-app/develop/server/basic-abilities/log-in/get-phone-number
func (d *Douyin) MiniLoginGetPhoneNumberInfo(code string) (*GetPhoneNumberResponse, *ErrorData) {
	getToken, err := d.ClientToken()
	if err != nil {
		return nil, NewErrorData(ECodeCall, err.Error())
	}
	req := &openApiSdkClient.V1GetPhonenumberInfoRequest{}
	req.SetAccessToken(getToken)
	req.SetCode(code)
	resp, err := d.client.V1GetPhonenumberInfo(req)
	if err != nil {
		var sdkError *tea.SDKError
		switch {
		case errors.As(err, &sdkError):
			errorCode, _ := strconv.Atoi(tea.StringValue(sdkError.Code))
			return nil, NewErrorData(errorCode, tea.StringValue(sdkError.Message))
		default:
			return nil, NewErrorData(ECodeCall, err.Error())
		}
	}
	if resp.Data == nil || *resp.Data == "" {
		if resp.ErrNo == nil {
			resp.ErrNo = tea.Int32(ECodeCall)
		}
		if resp.ErrMsg == nil {
			resp.ErrMsg = tea.String("调用失败")
		}
		return nil, NewErrorData(int(tea.Int32Value(resp.ErrNo)), tea.StringValue(resp.ErrMsg))
	}
	originText, err := d.RsaDecryptByPrivateKeyStr(*resp.Data)
	if err != nil {
		return nil, NewErrorData(ECodeCall, err.Error())
	}
	getPhoneNumberResponse := &GetPhoneNumberResponse{}
	err = json.Unmarshal([]byte(originText), getPhoneNumberResponse)
	if err != nil {
		return nil, NewErrorData(ECodeCall, err.Error())
	}
	return getPhoneNumberResponse, nil
}
