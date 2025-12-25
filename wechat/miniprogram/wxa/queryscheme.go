package wxa

import (
	"errors"
	"fmt"

	"resty.dev/v3"
)

type QuerySchemeRequest struct {
	Scheme    string `json:"scheme"`     // 小程序 scheme 码。支持加密 scheme 和明文 scheme
	QueryType int    `json:"query_type"` // 查询类型：查询 scheme 码信息：0， 查询每天剩余访问次数：1。（默认值0）
}

type QuerySchemeInfo struct {
	Appid      string `json:"appid"`       // 小程序 appid
	Path       string `json:"path"`        // 小程序页面路径
	Query      string `json:"query"`       // 小程序页面 query
	EnvVersion string `json:"env_version"` // 要打开的小程序版本：正式版为release 体验版为trial 开发版为develop
	ExpireTime int64  `json:"expire_time"` // 到期失效时间，为 Unix 时间戳，0 表示永久生效
	CreateTime int64  `json:"create_time"` // 创建时间，为 Unix 时间戳
}

type QuerySchemeQuotaInfo struct {
	RemainVisitQuota int64 `json:"remain_visit_quota"` // URL Scheme（加密+明文）/加密 URL Link 单天剩余访问次数
}

type QuerySchemeResponse struct {
	ErrCode    int                  `json:"errcode"`     // 错误码(0 成功)(-1 系统繁忙)(40001 无效access_token)(85403 scheme不存在)
	ErrMsg     string               `json:"errmsg"`      // 错误信息，请求失败时返回
	SchemeInfo QuerySchemeInfo      `json:"scheme_info"` // scheme 信息
	QuotaInfo  QuerySchemeQuotaInfo `json:"quota_info"`  // quota 配置
}

// QueryScheme 查询 scheme 码(queryScheme)
// DOC https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/qrcode-link/url-scheme/queryScheme.html
// POST https://api.weixin.qq.com/wxa/queryscheme?access_token=ACCESS_TOKEN
// BODY {"scheme":"xxx"}
func QueryScheme(httpClient *resty.Client, accessToken string, req *QuerySchemeRequest) (*QuerySchemeResponse, error) {
	httpResp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"access_token": accessToken,
		}).
		SetBody(req).
		SetResult(&QuerySchemeResponse{}).
		SetForceResponseContentType("application/json").
		Post("/wxa/queryscheme")
	if err != nil {
		return &QuerySchemeResponse{ErrCode: -1, ErrMsg: err.Error()}, err
	}
	if httpResp.IsSuccess() {
		resp, is := httpResp.Result().(*QuerySchemeResponse)
		if is {
			if resp.ErrCode != 0 {
				return resp, fmt.Errorf("%d %s", resp.ErrCode, resp.ErrMsg)
			}
			return resp, nil
		}
	}
	return &QuerySchemeResponse{
		ErrCode: httpResp.StatusCode(),
		ErrMsg:  httpResp.String(),
	}, errors.New(httpResp.String())
}
