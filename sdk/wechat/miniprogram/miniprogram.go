package miniprogram

import (
	"context"
	"sync"
	"time"

	"resty.dev/v3"

	"github.com/laixhe/gonet/sdk/wechat/apiutil"
	"github.com/laixhe/gonet/sdk/wechat/miniprogram/cgibin"
	"github.com/laixhe/gonet/sdk/wechat/miniprogram/sns"
	"github.com/laixhe/gonet/sdk/wechat/miniprogram/wxa"
)

// 微信小程序

// defaultHTTPTimeout 微信接口默认超时时间。
const defaultHTTPTimeout = 10 * time.Second

// ApiError 微信接口调用错误,可通过 errors.As 获取错误码。
//
//	var apiErr *miniprogram.ApiError
//	if errors.As(err, &apiErr) { _ = apiErr.ErrCode }
type ApiError = apiutil.ApiError

type Token struct {
	mutex       *sync.Mutex
	NetTime     int64  // 最新时间戳
	ExpiresIn   int64  // 凭证有效时间，单位：秒。目前是 7200 秒之内的值(2个小时)
	AccessToken string // 获取到的凭证
}

type MiniProgram struct {
	config     *Config
	httpClient *resty.Client
	token      *Token
}

func NewMiniProgram(config *Config) *MiniProgram {
	if err := config.Check(); err != nil {
		panic(err)
	}
	httpClient := resty.New()
	httpClient.SetBaseURL("https://api.weixin.qq.com")
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = defaultHTTPTimeout
	}
	httpClient.SetTimeout(timeout)
	return &MiniProgram{
		config:     config,
		httpClient: httpClient,
		token: &Token{
			mutex: &sync.Mutex{},
		},
	}
}

func (wx *MiniProgram) Config() *Config {
	return wx.config
}

// Code2Session 小程序登录
func (wx *MiniProgram) Code2Session(ctx context.Context, code string) (*sns.JsCode2SessionResponse, error) {
	return sns.JsCode2Session(ctx, wx.httpClient, wx.config.AppID, wx.config.Secret, code)
}

// GetAccessToken 获取接口调用凭据。
//
// forceRefresh 为 true 时忽略本地缓存,强制向微信刷新(可用于调试或令牌异常恢复)。
func (wx *MiniProgram) GetAccessToken(ctx context.Context, forceRefresh bool) (*cgibin.TokenResponse, error) {
	wx.token.mutex.Lock()
	defer wx.token.mutex.Unlock()

	if !forceRefresh && wx.token.NetTime > 0 && wx.token.ExpiresIn > 0 && wx.token.ExpiresIn > (time.Now().Unix()-wx.token.NetTime) {
		return wx.tokenResponse(), nil
	}
	tokenResp, err := cgibin.StableToken(ctx, wx.httpClient, wx.config.AppID, wx.config.Secret, forceRefresh)
	if err != nil {
		return nil, err
	}
	wx.token.AccessToken = tokenResp.AccessToken
	wx.token.NetTime = time.Now().Unix()
	if tokenResp.ExpiresIn > 200 {
		wx.token.ExpiresIn = tokenResp.ExpiresIn - 200
	} else {
		wx.token.ExpiresIn = 0
	}
	return wx.tokenResponse(), nil
}

// tokenResponse 返回缓存中的 token,命中缓存与刷新两条路径的返回值保持一致。
func (wx *MiniProgram) tokenResponse() *cgibin.TokenResponse {
	return &cgibin.TokenResponse{
		AccessToken: wx.token.AccessToken,
		ExpiresIn:   wx.token.ExpiresIn,
	}
}

// GetPhoneNumber 获取手机号
func (wx *MiniProgram) GetPhoneNumber(ctx context.Context, code string) (*wxa.GetUserPhoneNumberResponse, error) {
	getAccessToken, err := wx.GetAccessToken(ctx, false)
	if err != nil {
		return nil, err
	}
	return wxa.GetUserPhoneNumber(ctx, wx.httpClient, getAccessToken.AccessToken, code)
}

// GenerateScheme 获取加密 scheme 码
func (wx *MiniProgram) GenerateScheme(ctx context.Context, req *wxa.GenerateSchemeRequest) (*wxa.GenerateSchemeResponse, error) {
	getAccessToken, err := wx.GetAccessToken(ctx, false)
	if err != nil {
		return nil, err
	}
	return wxa.GenerateScheme(ctx, wx.httpClient, getAccessToken.AccessToken, req)
}

// QueryScheme 查询 scheme 码
func (wx *MiniProgram) QueryScheme(ctx context.Context, req *wxa.QuerySchemeRequest) (*wxa.QuerySchemeResponse, error) {
	getAccessToken, err := wx.GetAccessToken(ctx, false)
	if err != nil {
		return nil, err
	}
	return wxa.QueryScheme(ctx, wx.httpClient, getAccessToken.AccessToken, req)
}

// GetWxaCodeUnlimit 获取不限制的小程序码
func (wx *MiniProgram) GetWxaCodeUnlimit(ctx context.Context, req *wxa.GetWxaCodeUnlimitRequest) (*wxa.GetWxaCodeUnlimitResponse, error) {
	getAccessToken, err := wx.GetAccessToken(ctx, false)
	if err != nil {
		return nil, err
	}
	return wxa.GetWxaCodeUnlimit(ctx, wx.httpClient, getAccessToken.AccessToken, req)
}
