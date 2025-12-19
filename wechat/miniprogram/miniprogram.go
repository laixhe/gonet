package miniprogram

import (
	"sync"
	"time"

	"resty.dev/v3"

	"github.com/laixhe/gonet/wechat/miniprogram/cgibin"
	"github.com/laixhe/gonet/wechat/miniprogram/sns"
	"github.com/laixhe/gonet/wechat/miniprogram/wxa"
)

type Token struct {
	mutex       *sync.Mutex
	NetTime     int64  // 最新时间戳
	ExpiresIn   int64  // 凭证有效时间，单位：秒。目前是 7200 秒之内的值(2个小时)
	AccessToken string // 获取到的凭证
}

// MiniProgram 微信小程序
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
	return &MiniProgram{
		config:     config,
		httpClient: httpClient,
		token: &Token{
			mutex: &sync.Mutex{},
		},
	}
}

// Config 获取配置
func (wx *MiniProgram) Config() *Config {
	return wx.config
}

// Code2Session 小程序登录
func (wx *MiniProgram) Code2Session(code string) (*sns.JsCode2SessionResponse, error) {
	return sns.JsCode2Session(wx.httpClient, wx.config.AppId, wx.config.Secret, code)
}

// GetAccessToken 获取接口调用凭据
func (wx *MiniProgram) GetAccessToken() (*cgibin.TokenResponse, error) {
	wx.token.mutex.Lock()
	defer wx.token.mutex.Unlock()

	if wx.token.NetTime > 0 && wx.token.ExpiresIn > 0 && wx.token.ExpiresIn > (time.Now().Unix()-wx.token.NetTime) {
		return &cgibin.TokenResponse{
			AccessToken: wx.token.AccessToken,
			ExpiresIn:   wx.token.ExpiresIn,
		}, nil
	}
	token, err := cgibin.StableToken(wx.httpClient, wx.config.AppId, wx.config.Secret, false)
	if err != nil {
		return nil, err
	}
	wx.token.AccessToken = token.AccessToken
	wx.token.NetTime = time.Now().Unix()
	if token.ExpiresIn > 300 {
		wx.token.ExpiresIn = token.ExpiresIn - 300
	} else {
		wx.token.ExpiresIn = 0
	}
	return token, nil
}

// GetPhoneNumber 获取手机号
func (wx *MiniProgram) GetPhoneNumber(code string) (*wxa.GetUserPhoneNumberResponse, error) {
	getAccessToken, err := wx.GetAccessToken()
	if err != nil {
		return nil, err
	}
	return wxa.GetUserPhoneNumber(wx.httpClient, getAccessToken.AccessToken, code)
}

// GenerateScheme 获取加密 scheme 码
func (wx *MiniProgram) GenerateScheme(req *wxa.GenerateSchemeRequest) (*wxa.GenerateSchemeResponse, error) {
	getAccessToken, err := wx.GetAccessToken()
	if err != nil {
		return nil, err
	}
	return wxa.GenerateScheme(wx.httpClient, getAccessToken.AccessToken, req)
}

// QueryScheme 查询 scheme 码
func (wx *MiniProgram) QueryScheme(req *wxa.QuerySchemeRequest) (*wxa.QuerySchemeResponse, error) {
	getAccessToken, err := wx.GetAccessToken()
	if err != nil {
		return nil, err
	}
	return wxa.QueryScheme(wx.httpClient, getAccessToken.AccessToken, req)
}

// GetWxaCodeUnlimit 获取不限制的小程序码
func (wx *MiniProgram) GetWxaCodeUnlimit(req *wxa.GetWxaCodeUnlimitRequest) (*wxa.GetWxaCodeUnlimitResponse, error) {
	getAccessToken, err := wx.GetAccessToken()
	if err != nil {
		return nil, err
	}
	return wxa.GetWxaCodeUnlimit(wx.httpClient, getAccessToken.AccessToken, req)
}
