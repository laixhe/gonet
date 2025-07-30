package sms

import (
	"encoding/json"
	"errors"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	"github.com/alibabacloud-go/tea/tea"
)

/*
alibaba_sms:
  access_key_id: XXXXXX
  access_key_secret: XXXXXX
  endpoint: dysmsapi.aliyuncs.com
  sends:
    - key: XXXXXX通知
      sign_name: XXXXXX网络
      template_code: SMS_XXXXXX
    - key: XXXXXX用户通知
      sign_name: XXXXXX网络
      template_code: SMS_XXXXXX
*/

// ConfigSend 短信发送配置
type ConfigSend struct {
	// 短信模板名称
	Key string `json:"key" mapstructure:"key" toml:"key" yaml:"key"`
	// 短信签名名称
	SignName string `json:"sign_name" mapstructure:"sign_name" toml:"sign_name" yaml:"sign_name"`
	// 短信模板CODE
	TemplateCode string `json:"template_code" mapstructure:"template_code" toml:"template_code" yaml:"template_code"`
}

// Config 短信配置
type Config struct {
	// 访问密钥ID
	AccessKeyId string `json:"access_key_id" mapstructure:"access_key_id" toml:"access_key_id" yaml:"access_key_id"`
	// 访问密钥
	AccessKeySecret string `json:"access_key_secret" mapstructure:"access_key_secret" toml:"access_key_secret" yaml:"access_key_secret"`
	// 访问域名
	Endpoint string `json:"endpoint" mapstructure:"endpoint" toml:"endpoint" yaml:"endpoint"`
	// 短信发送配置
	Sends []ConfigSend `json:"sends" mapstructure:"sends" toml:"sends" yaml:"sends"`
}

// Checking 检查
func (c *Config) Check() error {
	if c == nil {
		return errors.New("没有短信配置")
	}
	if c.AccessKeyId == "" {
		return errors.New("没有短信访问密钥ID配置")
	}
	if c.AccessKeySecret == "" {
		return errors.New("没有短信访问密钥配置")
	}
	if c.Endpoint == "" {
		return errors.New("没有短信访问域名配置")
	}
	if len(c.Sends) == 0 {
		return errors.New("没有短信发送配置")
	}
	for k, v := range c.Sends {
		if v.Key == "" {
			return fmt.Errorf("短信发送配置[%d]没有名称", k)
		}
		if v.SignName == "" {
			return fmt.Errorf("短信发送配置[%d]没有签名名称", k)
		}
		if v.TemplateCode == "" {
			return fmt.Errorf("短信发送配置[%d]没有模板CODE", k)
		}
	}
	return nil
}

type SmsClient struct {
	config *Config
	client *dysmsapi.Client
}

// Send 发送短信
// key 短信发送配置名称
// phone 手机号
// params 短信模板参数
// 返回值：短信发送ID，错误信息
func (s *SmsClient) Send(key string, phone string, params map[string]string) (string, error) {
	templateParam := ""
	if len(params) > 0 {
		paramsJson, err := json.Marshal(params)
		if err != nil {
			return "", err
		}
		templateParam = string(paramsJson)
	}
	for _, v := range s.config.Sends {
		if v.Key == key {
			request := &dysmsapi.SendSmsRequest{}
			request.SetPhoneNumbers(phone)
			request.SetSignName(v.SignName)
			request.SetTemplateCode(v.TemplateCode)
			if len(templateParam) > 0 {
				request.SetTemplateParam(templateParam)
			}
			response, err := s.client.SendSms(request)
			if err != nil {
				return "", err
			}
			if response.Body == nil {
				return "", errors.New("短信发送失败")
			}
			if response.Body.Code != nil && *response.Body.Code != "OK" {
				return "", errors.New("短信发送失败：" + tea.StringValue(response.Body.Message))
			}
			return tea.StringValue(response.Body.BizId), nil
		}
	}
	return "", errors.New("没有短信发送配置：" + key)
}

func Init(config *Config) (*SmsClient, error) {
	if err := config.Check(); err != nil {
		return nil, err
	}
	openapiConfig := &openapi.Config{
		AccessKeyId:     tea.String(config.AccessKeyId),
		AccessKeySecret: tea.String(config.AccessKeySecret),
	}
	openapiConfig.Endpoint = tea.String(config.Endpoint)
	client, err := dysmsapi.NewClient(openapiConfig)
	if err != nil {
		return nil, err
	}
	return &SmsClient{
		config: config,
		client: client,
	}, nil
}
