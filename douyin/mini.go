package douyin

import (
	"github.com/alibabacloud-go/tea/tea"
	openApiSdkClient "github.com/bytedance/douyin-openapi-sdk-go/client"
)

// MiniJscode2session 小程序登录
// 通常用于小程序通过临时登录凭证 code 换取用户唯一标识 openid 、session_key 等信息
// DOC https://developer.open-douyin.com/docs/resource/zh-CN/mini-app/develop/server/basic-abilities/log-in/code-2-session
func (d *Douyin) MiniJscode2session(code string, anonymousCode ...string) *openApiSdkClient.V2Jscode2sessionResponse {
	req := &openApiSdkClient.V2Jscode2sessionRequest{}
	req.SetAppid(d.config.AppID)
	req.SetSecret(d.config.AppSecret)
	req.SetCode(code)
	if len(anonymousCode) > 0 {
		req.SetAnonymousCode(anonymousCode[0])
	}
	resp, err := d.client.V2Jscode2session(req)
	if err != nil {
		return &openApiSdkClient.V2Jscode2sessionResponse{
			ErrTips: tea.String(err.Error()),
			Data:    &openApiSdkClient.V2Jscode2sessionResponseData{},
			LogId:   tea.String(""),
			ErrNo:   tea.Int64(-1),
		}
	}
	return resp
}
