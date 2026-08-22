package sms

import (
	"encoding/json"
	"errors"

	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	"github.com/alibabacloud-go/tea/tea"
)

// Send 发送短信
// key 短信发送配置名称
// phone 手机号
// params 短信模板参数
// 返回值：短信发送ID，错误信息
func (sc *SClient) Send(key string, phone string, params map[string]string) (*dysmsapi.SendSmsResponseBody, error) {
	templateParam := ""
	if len(params) > 0 {
		paramsJson, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		templateParam = string(paramsJson)
	}
	for _, v := range sc.config.Sends {
		if v.Key == key {
			request := &dysmsapi.SendSmsRequest{}
			request.SetPhoneNumbers(phone)
			request.SetSignName(v.SignName)
			request.SetTemplateCode(v.TemplateCode)
			if len(templateParam) > 0 {
				request.SetTemplateParam(templateParam)
			}
			response, err := sc.client.SendSms(request)
			if err != nil {
				return nil, err
			}
			if err := checkSendSmsResp(response.Body); err != nil {
				return nil, err
			}
			return response.Body, nil
		}
	}
	return nil, errors.New("没有短信发送配置：" + key)
}

// checkSendSmsResp 校验短信发送响应,成功返回 nil
// Code 为空(响应体畸形)或非 "OK" 均视为失败
func checkSendSmsResp(body *dysmsapi.SendSmsResponseBody) error {
	if body == nil {
		return errors.New("短信发送失败")
	}
	if body.Code == nil || *body.Code != "OK" {
		return errors.New("短信发送失败：" + tea.StringValue(body.Message))
	}
	return nil
}
