package wxa

import (
	"errors"
	"fmt"

	"github.com/bytedance/sonic"
	"resty.dev/v3"
)

type GenerateSchemeJumpWxa struct {
	Path       string `json:"path,omitempty"`        // 通过 scheme 码进入的小程序页面路径，必须是已经发布的小程序存在的页面，不可携带 query，path 为空时会跳转小程序主页
	Query      string `json:"query,omitempty"`       // 通过 scheme 码进入小程序时的 query，最大1024个字符，只支持数字，大小写英文以及部分特殊字 = &（格式遵循URL标准，即k1=v1&k2=v2）
	EnvVersion string `json:"env_version,omitempty"` // 默认值 release 要打开的小程序版本。正式版为release，体验版为trial，开发版为develop，仅在微信外打开时生效
}
type GenerateSchemeRequest struct {
	JumpWxa        GenerateSchemeJumpWxa `json:"jump_wxa,omitempty"`        // 跳转到的目标小程序信息
	IsExpire       bool                  `json:"is_expire,omitempty"`       // 是否开启到期失效时间
	ExpireType     int64                 `json:"expire_type,omitempty"`     // 默认值：0，到期失效的 scheme 码失效类型，失效时间：0，失效间隔天数：1
	ExpireTime     int64                 `json:"expire_time,omitempty"`     // 到期失效的 scheme 码的失效时间，为 Unix 时间戳。生成的到期失效 scheme 码在该时间前有效。最长有效期为30天。is_expire 为 true 且 expire_type 为 0 时必填
	ExpireInterval int64                 `json:"expire_interval,omitempty"` // 到期失效的 scheme 码的失效间隔天数。生成的到期失效 scheme 码在该间隔时间到达前有效。最长间隔天数为30天。is_expire 为 true 且 expire_type 为 1 时必填
}

type GenerateSchemeResponse struct {
	ErrCode  int    `json:"errcode"`  // 错误码(0 成功)(-1 系统繁忙)(40001 无效access_token)(85406 单天累加访问次数超过上限)
	ErrMsg   string `json:"errmsg"`   // 错误信息，请求失败时返回
	OpenLink string `json:"openlink"` // 生成的小程序 scheme 码
}

// GenerateScheme 获取加密 scheme 码(generateScheme)
// DOC https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/qrcode-link/url-scheme/generateScheme.html
// POST https://api.weixin.qq.com/wxa/generatescheme?access_token=ACCESS_TOKEN
// BODY {"jump_wxa":{"path":"/pages/index/index","query":"id=1&age=18"}}
func GenerateScheme(httpClient *resty.Client, accessToken string, req *GenerateSchemeRequest) (*GenerateSchemeResponse, error) {
	// 原生的 json 会对类似的字符串路径的 / 进行转义，如 {"path": "xxx/xxx/xxx"}
	reqBody, err := sonic.Marshal(req)
	if err != nil {
		return &GenerateSchemeResponse{
			ErrCode: 400,
			ErrMsg:  err.Error(),
		}, err
	}
	httpResp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"access_token": accessToken,
		}).
		SetBody(string(reqBody)).
		SetResult(&GenerateSchemeResponse{}).
		SetForceResponseContentType("application/json").
		Post("/wxa/generatescheme")
	if err != nil {
		return &GenerateSchemeResponse{ErrCode: -1, ErrMsg: err.Error()}, err
	}
	if httpResp.IsSuccess() {
		resp, is := httpResp.Result().(*GenerateSchemeResponse)
		if is {
			if resp.ErrCode != 0 {
				return resp, fmt.Errorf("%d %s", resp.ErrCode, resp.ErrMsg)
			}
			return resp, nil
		}
	}
	return &GenerateSchemeResponse{
		ErrCode: httpResp.StatusCode(),
		ErrMsg:  httpResp.String(),
	}, errors.New(httpResp.String())
}
