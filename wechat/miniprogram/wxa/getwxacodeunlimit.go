package wxa

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
	"resty.dev/v3"
)

type GetWxaCodeUnlimitRequest struct {
	Page       string `json:"page"`                  // 页面 page 例如 pages/index/index 根路径前不要填加 / 不能携带参数（参数请放在scene字段里），如果不填写这个字段，默认跳主页面
	Scene      string `json:"scene"`                 // 通过小程序码进入小程序时的 query 最大1024个字符，只支持数字，大小写英文以及部分特殊字 = &（格式遵循URL标准，即k1=v1&k2=v2）
	CheckPath  bool   `json:"check_path"`            // 检查 page 是否存在，为 true 时 page 必须是已经发布的小程序存在的页面（否则报错）；为 false 时允许小程序未发布或者 page 不存在， 但 page 有数量上限（60000个）请勿滥用
	EnvVersion string `json:"env_version,omitempty"` // 默认值 release 要打开的小程序版本。正式版为release，体验版为trial，开发版为develop，仅在微信外打开时生效
	Width      int    `json:"width,omitempty"`       // 默认 430 二维码的宽度，单位 px，最小 280 px，最大 1280 px
	IsHyaline  bool   `json:"is_hyaline,omitempty"`  // 默认是 false 是否需要透明底色，为 true 时，生成透明底色的小程序
}

type GetWxaCodeUnlimitResponse struct {
	ErrCode     int    `json:"errcode"`      // 错误信息，请求失败时返回(-1 系统繁忙)(40001 无效access_token)(40129 scene参数不正确)(41030 page路径不正确)
	ErrMsg      string `json:"errmsg"`       // 错误码，请求失败时返回
	ContentType string `json:"content_type"` // 图片响应类型
	Buffer      []byte `json:"buffer"`       // 图片 Buffer
}

// GetWxaCodeUnlimit 获取不限制的小程序码(getUnlimitedQRCode)
// DOC https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/qrcode-link/qr-code/getUnlimitedQRCode.html
// POST https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=ACCESS_TOKEN
// BODY {"page":"xxx","scene":"xxx","width":1280}
func GetWxaCodeUnlimit(httpClient *resty.Client, accessToken string, req *GetWxaCodeUnlimitRequest) (*GetWxaCodeUnlimitResponse, error) {
	if req.Width <= 0 {
		req.Width = 1280
	}
	reqBody, err := sonic.Marshal(req)
	if err != nil {
		return &GetWxaCodeUnlimitResponse{
			ErrCode: 400,
			ErrMsg:  err.Error(),
		}, err
	}
	httpResp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"access_token": accessToken,
		}).
		SetBody(string(reqBody)).
		Post("/wxa/getwxacodeunlimit")
	if err != nil {
		return &GetWxaCodeUnlimitResponse{ErrCode: -1, ErrMsg: err.Error()}, err
	}
	if httpResp.IsSuccess() {
		data := httpResp.Bytes()
		contentType := httpResp.Header().Get("Content-Type")
		if strings.HasPrefix(contentType, "image") {
			return &GetWxaCodeUnlimitResponse{
				ContentType: contentType,
				Buffer:      data,
			}, nil
		}
		resp := &GetWxaCodeUnlimitResponse{}
		if err = json.Unmarshal(data, resp); err != nil {
			resp.ErrCode = 500
			resp.ErrMsg = err.Error()
		}
		return resp, fmt.Errorf("%d %s", resp.ErrCode, resp.ErrMsg)
	}
	return &GetWxaCodeUnlimitResponse{
		ErrCode: httpResp.StatusCode(),
		ErrMsg:  httpResp.String(),
	}, errors.New(httpResp.String())
}
