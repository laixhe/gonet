package douyin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	openApiSdkClient "github.com/bytedance/douyin-openapi-sdk-go/client"
)

// 内容安全检测
// 检测一段文本是否包含违法违规内容
// DOC https://developer.open-douyin.com/docs/resource/zh-CN/mini-app/develop/server/basic-abilities/content-security/content-security-detect-new

const (
	// textCensorURL 内容安全检测线上地址,可通过 Config.TextCensorURL 覆盖(如沙箱地址)
	textCensorURL = "https://open.douyin.com/api/apps/v1/censor/text/"
	// contentSecurityTimeout 内容安全检测请求超时时间
	contentSecurityTimeout = 10 * time.Second
)

// TextCensorRequest 内容安全检测请求
type TextCensorRequest struct {
	AppID   string   `json:"app_id"`  // 小程序 app_id
	Content []string `json:"content"` // 检测的文本内容列表
}

// TextCensorPredict 检测结果-置信度
type TextCensorPredict struct {
	Target    string  `json:"target"`     // 服务/目标
	ModelName string  `json:"model_name"` // 模型/标签
	Prob      float64 `json:"prob"`       // 概率，仅供参考，可以忽略
	Hit       bool    `json:"hit"`        // 结果，为 true 时表示检测的文本包含违法违规内容
}

// TextCensorResult 检测结果
type TextCensorResult struct {
	Code     int                 `json:"code"`     // 检测结果-状态码，0 表示成功
	Msg      string              `json:"msg"`      // 检测结果-消息
	DataID   string              `json:"data_id"`  // 检测结果-数据 id
	TaskID   string              `json:"task_id"`  // 检测结果-任务 id
	Predicts []TextCensorPredict `json:"predicts"` // 检测结果-置信度列表
}

// TextCensorResponse 内容安全检测响应
type TextCensorResponse struct {
	LogID  string             `json:"log_id"`  // 请求 id
	Data   []TextCensorResult `json:"data"`    // 检测结果列表
	ErrNo  int64              `json:"err_no"`  // 请求级状态码，0 表示成功
	ErrMsg string             `json:"err_msg"` // 请求级错误信息
}

// ContentSecurityTextDetect 内容安全检测
// 检测一段或多段文本是否包含违法违规内容，返回原始检测结果
// 注意：需要在小程序后台开通能力「检测文本是否包含违法违规内容V2」
// 沙箱联调：配置 Config.TextCensorURL 指向沙箱地址；自定义超时/代理可通过 SetTextCensorHTTPClient 注入
func (d *Douyin) ContentSecurityTextDetect(ctx context.Context, contents ...string) (*TextCensorResponse, error) {
	if len(contents) == 0 {
		return nil, newError(ErrKindLocal, ECodeCall, 0, "检测内容不能为空")
	}
	for _, content := range contents {
		if content == "" {
			return nil, newError(ErrKindLocal, ECodeCall, 0, "检测内容不能为空")
		}
	}
	// 获取 client_token（接口 access-token 头传 client_token）
	getToken, err := d.ClientToken()
	if err != nil {
		return nil, err
	}
	url := textCensorURL
	if d.config.TextCensorURL != "" {
		url = d.config.TextCensorURL
	}
	body, err := json.Marshal(&TextCensorRequest{
		AppID:   d.config.AppID,
		Content: contents,
	})
	if err != nil {
		return nil, newError(ErrKindLocal, ECodeCall, 0, err.Error())
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, newError(ErrKindLocal, ECodeCall, 0, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("access-token", getToken)
	httpClient := d.textCensorClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: contentSecurityTimeout}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, newError(ErrKindNetwork, ECodeCall, 0, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, newError(ErrKindHTTP, ECodeCall, resp.StatusCode, fmt.Sprintf("内容安全检测 HTTP 状态码 %d", resp.StatusCode))
	}
	textCensorResponse := &TextCensorResponse{}
	if err = json.NewDecoder(resp.Body).Decode(textCensorResponse); err != nil {
		return nil, newError(ErrKindDecode, ECodeCall, 0, err.Error())
	}
	// 请求级失败（err_no 非 0，如 access_token 无效/能力未开通等）
	if textCensorResponse.ErrNo != 0 {
		return nil, newError(ErrKindBusiness, int(textCensorResponse.ErrNo), 0, textCensorResponse.ErrMsg)
	}
	// 任务级失败
	for _, item := range textCensorResponse.Data {
		if item.Code != 0 {
			return nil, newError(ErrKindBusiness, item.Code, 0, item.Msg)
		}
	}
	return textCensorResponse, nil
}

// ContentSecurityTextSafe 内容安全检测，返回文本是否安全
// 安全返回 true；包含违法违规内容返回 false
func (d *Douyin) ContentSecurityTextSafe(ctx context.Context, contents ...string) (bool, error) {
	textCensorResponse, err := d.ContentSecurityTextDetect(ctx, contents...)
	if err != nil {
		return false, err
	}
	for _, item := range textCensorResponse.Data {
		for _, predict := range item.Predicts {
			if predict.Hit {
				return false, nil
			}
		}
	}
	return true, nil
}

// ==================== 图片内容安全检测 V3 ====================
// 检测一张图片是否包含违法违规内容
// DOC https://developer.open-douyin.com/docs/resource/zh-CN/mini-app/develop/server/basic-abilities/content-security/picture-detect-v3

// ImageCensorPredict 检测结果-置信度
type ImageCensorPredict struct {
	ModelName string `json:"model_name"` // 模型/标签
	Hit       bool   `json:"hit"`        // 结果，为 true 时表示检测的图片包含违法违规内容
}

// ImageCensorResponse 图片内容安全检测响应
type ImageCensorResponse struct {
	LogID    string               `json:"log_id"`   // 请求 id
	Predicts []ImageCensorPredict `json:"predicts"` // 检测结果-置信度列表
	ErrNo    int64                `json:"err_no"`   // 请求级状态码，0 表示成功
	ErrMsg   string               `json:"err_msg"`  // 请求级错误信息
}

// ContentSecurityImageDetect 图片内容安全检测 V3
// 通过图片链接或 base64 数据检测图片是否包含违法违规内容，返回原始检测结果
// image 与 imageData 至少传一个
// 注意：需要在小程序后台开通能力「数据安全」或「检测图片是否包含违法违规内容V2」
func (d *Douyin) ContentSecurityImageDetect(image, imageData string) (*ImageCensorResponse, error) {
	if image == "" && imageData == "" {
		return nil, newError(ErrKindLocal, ECodeCall, 0, "图片链接和图片 base64 数据不能同时为空")
	}
	// 获取 client_token（接口 access-token 头传 client_token）
	getToken, err := d.ClientToken()
	if err != nil {
		return nil, err
	}
	// 复用官方 SDK 的 CensorImage
	req := &openApiSdkClient.CensorImageRequest{}
	req.SetAppId(d.config.AppID)
	req.SetImage(image)
	req.SetImageData(imageData)
	req.SetAccessToken(getToken)
	resp, err := d.client.CensorImage(req)
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
	// 请求级业务错误(err_no 非 0)优先返回,与 ContentSecurityTextDetect 行为对齐
	if resp.ErrNo != nil && tea.Int32Value(resp.ErrNo) != 0 {
		return nil, newError(ErrKindBusiness, int(tea.Int32Value(resp.ErrNo)), 0, tea.StringValue(resp.ErrMsg))
	}
	imageCensorResponse := &ImageCensorResponse{
		LogID:  tea.StringValue(resp.LogId),
		ErrNo:  int64(tea.Int32Value(resp.ErrNo)),
		ErrMsg: tea.StringValue(resp.ErrMsg),
	}
	imageCensorResponse.Predicts = make([]ImageCensorPredict, 0, len(resp.Predicts))
	for _, item := range resp.Predicts {
		imageCensorResponse.Predicts = append(imageCensorResponse.Predicts, ImageCensorPredict{
			ModelName: tea.StringValue(item.ModelName),
			Hit:       tea.BoolValue(item.Hit),
		})
	}
	return imageCensorResponse, nil
}

// ContentSecurityImageSafe 图片内容安全检测 V3，返回图片是否安全
// 安全返回 true；包含违法违规内容返回 false
func (d *Douyin) ContentSecurityImageSafe(image, imageData string) (bool, error) {
	imageCensorResponse, err := d.ContentSecurityImageDetect(image, imageData)
	if err != nil {
		return false, err
	}
	for _, predict := range imageCensorResponse.Predicts {
		if predict.Hit {
			return false, nil
		}
	}
	return true, nil
}
