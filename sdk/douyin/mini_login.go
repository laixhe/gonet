package douyin

import (
	"encoding/json"
	"errors"

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
func (d *Douyin) MiniLoginJsCode2Session(code string, anonymousCode ...string) (*JsCode2SessionResponse, error) {
	anonymous := ""
	if len(anonymousCode) > 0 {
		anonymous = anonymousCode[0]
	}
	if code == "" && anonymous == "" {
		return nil, newError(ErrKindLocal, ECodeCall, 0, "code 与 anonymous_code 不能同时为空")
	}
	req := &openApiSdkClient.V2Jscode2sessionRequest{}
	req.SetAppid(d.config.AppID)
	req.SetSecret(d.config.AppSecret)
	if code != "" {
		req.SetCode(code)
	} else {
		req.SetAnonymousCode(anonymous)
	}
	resp, err := d.client.V2Jscode2session(req)
	if err != nil {
		var sdkError *tea.SDKError
		switch {
		case errors.As(err, &sdkError):
			errCode, kind := parseSDKErrorCode(sdkError)
			return nil, newError(kind, errCode, 0, tea.StringValue(sdkError.Message))
		default:
			return nil, newError(ErrKindLocal, ECodeCall, 0, err.Error())
		}
	}
	return parseJsCode2Session(resp, code, anonymous)
}

// parseJsCode2Session 校验并转换 code2session 响应。
// 注意:unionid 需在开发者后台绑定后才返回,未绑定的小程序所有用户 unionid 均为空,
// 因此 code 模式仅强制校验 openid(登录主键),unionid 允许为空。
func parseJsCode2Session(resp *openApiSdkClient.V2Jscode2sessionResponse, code, anonymous string) (*JsCode2SessionResponse, error) {
	if resp == nil {
		return nil, newError(ErrKindLocal, ECodeCall, 0, "调用失败:响应为空")
	}
	// 请求级业务错误(err_no 非 0)优先返回,避免 data 非空时误判成功
	if resp.ErrNo != nil && tea.Int64Value(resp.ErrNo) != 0 {
		return nil, newError(ErrKindBusiness, int(tea.Int64Value(resp.ErrNo)), 0, tea.StringValue(resp.ErrTips))
	}
	if code != "" {
		if resp.Data == nil || resp.Data.Openid == nil || *resp.Data.Openid == "" {
			return nil, jsCode2SessionError(resp.ErrNo, resp.ErrTips)
		}
	}
	if code == "" && anonymous != "" {
		if resp.Data == nil || resp.Data.AnonymousOpenid == nil || *resp.Data.AnonymousOpenid == "" {
			return nil, jsCode2SessionError(resp.ErrNo, resp.ErrTips)
		}
	}
	return &JsCode2SessionResponse{
		OpenID:          tea.StringValue(resp.Data.Openid),
		UnionID:         tea.StringValue(resp.Data.Unionid),
		SessionKey:      tea.StringValue(resp.Data.SessionKey),
		AnonymousOpenID: tea.StringValue(resp.Data.AnonymousOpenid),
	}, nil
}

// jsCode2SessionError 兜底构造业务错误:err_no/err_tips 缺失时回退为通用失败
func jsCode2SessionError(errNo *int64, errTips *string) error {
	if errNo == nil {
		errNo = tea.Int64(ECodeCall)
	}
	if errTips == nil {
		errTips = tea.String("调用失败")
	}
	return newError(ErrKindBusiness, int(tea.Int64Value(errNo)), 0, tea.StringValue(errTips))
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
func (d *Douyin) MiniLoginGetPhoneNumberInfo(code string) (*GetPhoneNumberResponse, error) {
	getToken, err := d.ClientToken()
	if err != nil {
		return nil, err
	}
	req := &openApiSdkClient.V1GetPhonenumberInfoRequest{}
	req.SetAccessToken(getToken)
	req.SetCode(code)
	resp, err := d.client.V1GetPhonenumberInfo(req)
	if err != nil {
		var sdkError *tea.SDKError
		switch {
		case errors.As(err, &sdkError):
			errCode, kind := parseSDKErrorCode(sdkError)
			return nil, newError(kind, errCode, 0, tea.StringValue(sdkError.Message))
		default:
			return nil, newError(ErrKindLocal, ECodeCall, 0, err.Error())
		}
	}
	if resp == nil {
		return nil, newError(ErrKindLocal, ECodeCall, 0, "调用失败:响应为空")
	}
	// 请求级业务错误(err_no 非 0)优先返回,避免 data 非空时误判成功
	if resp.ErrNo != nil && tea.Int32Value(resp.ErrNo) != 0 {
		return nil, newError(ErrKindBusiness, int(tea.Int32Value(resp.ErrNo)), 0, tea.StringValue(resp.ErrMsg))
	}
	if resp.Data == nil || *resp.Data == "" {
		if resp.ErrNo == nil {
			resp.ErrNo = tea.Int32(ECodeCall)
		}
		if resp.ErrMsg == nil {
			resp.ErrMsg = tea.String("调用失败")
		}
		return nil, newError(ErrKindBusiness, int(tea.Int32Value(resp.ErrNo)), 0, tea.StringValue(resp.ErrMsg))
	}
	originText, err := d.RsaDecryptByPrivateKeyStr(*resp.Data)
	if err != nil {
		return nil, newError(ErrKindLocal, ECodeCall, 0, err.Error())
	}
	getPhoneNumberResponse := &GetPhoneNumberResponse{}
	err = json.Unmarshal([]byte(originText), getPhoneNumberResponse)
	if err != nil {
		return nil, newError(ErrKindDecode, ECodeCall, 0, err.Error())
	}
	return getPhoneNumberResponse, nil
}
