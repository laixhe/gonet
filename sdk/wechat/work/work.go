package work

import (
	"context"
	"sync"
	"time"

	"resty.dev/v3"

	"github.com/laixhe/gonet/sdk/wechat/apiutil"
	"github.com/laixhe/gonet/sdk/wechat/work/cgibin"
)

// 企业微信

// defaultHTTPTimeout 企业微信接口默认超时时间。
const defaultHTTPTimeout = 10 * time.Second

// ApiError 企业微信接口调用错误,可通过 errors.As 获取错误码。
//
//	var apiErr *work.ApiError
//	if errors.As(err, &apiErr) { _ = apiErr.ErrCode }
type ApiError = apiutil.ApiError

type Token struct {
	mutex       *sync.Mutex
	NetTime     int64  // 最新时间戳
	ExpiresIn   int64  // 凭证有效时间，单位：秒。目前是 7200 秒之内的值(2个小时)
	AccessToken string // 获取到的凭证
}

type Work struct {
	config     *Config
	httpClient *resty.Client
	token      *Token
}

func NewWork(config *Config) *Work {
	if err := config.Check(); err != nil {
		panic(err)
	}
	httpClient := resty.New()
	httpClient.SetBaseURL("https://qyapi.weixin.qq.com")
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = defaultHTTPTimeout
	}
	httpClient.SetTimeout(timeout)
	return &Work{
		config:     config,
		httpClient: httpClient,
		token: &Token{
			mutex: &sync.Mutex{},
		},
	}
}

func (w *Work) Config() *Config {
	return w.config
}

// GetToken 获取接口调用凭据。
//
// forceRefresh 为 true 时忽略本地缓存,强制向微信刷新(可用于调试或令牌异常恢复)。
func (w *Work) GetToken(ctx context.Context, forceRefresh bool) (*cgibin.GetTokenResponse, error) {
	w.token.mutex.Lock()
	defer w.token.mutex.Unlock()

	if !forceRefresh && w.token.NetTime > 0 && w.token.ExpiresIn > 0 && w.token.ExpiresIn > (time.Now().Unix()-w.token.NetTime) {
		return w.tokenResponse(), nil
	}
	tokenResp, err := cgibin.GetToken(ctx, w.httpClient, w.config.CorpID, w.config.CorpSecret)
	if err != nil {
		return nil, err
	}
	w.token.AccessToken = tokenResp.AccessToken
	w.token.NetTime = time.Now().Unix()
	if tokenResp.ExpiresIn > 200 {
		w.token.ExpiresIn = tokenResp.ExpiresIn - 200
	} else {
		w.token.ExpiresIn = 0
	}
	return w.tokenResponse(), nil
}

// tokenResponse 返回缓存中的 token,命中缓存与刷新两条路径的返回值保持一致。
func (w *Work) tokenResponse() *cgibin.GetTokenResponse {
	return &cgibin.GetTokenResponse{
		AccessToken: w.token.AccessToken,
		ExpiresIn:   w.token.ExpiresIn,
	}
}

// GetUserInfo 获取访问用户身份
func (w *Work) GetUserInfo(ctx context.Context, code string) (*cgibin.GetUserInfoResponse, error) {
	getToken, err := w.GetToken(ctx, false)
	if err != nil {
		return nil, err
	}
	return cgibin.GetUserInfo(ctx, w.httpClient, getToken.AccessToken, code)
}

// GetUserDetail 获取访问用户敏感信息
func (w *Work) GetUserDetail(ctx context.Context, userTicket string) (*cgibin.GetUserDetailResponse, error) {
	getToken, err := w.GetToken(ctx, false)
	if err != nil {
		return nil, err
	}
	return cgibin.GetUserDetail(ctx, w.httpClient, getToken.AccessToken, userTicket)
}

// UserGet 读取成员
func (w *Work) UserGet(ctx context.Context, userID string) (*cgibin.UserGetResponse, error) {
	getToken, err := w.GetToken(ctx, false)
	if err != nil {
		return nil, err
	}
	return cgibin.UserGet(ctx, w.httpClient, getToken.AccessToken, userID)
}
