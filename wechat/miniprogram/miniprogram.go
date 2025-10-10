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
	AccessToken string // 获取到的凭证
	ExpiresIn   int64  // 凭证有效时间，单位：秒。目前是7200秒之内的值(2个小时)
}

// Miniprogram 微信小程序
type Miniprogram struct {
	config     *Config
	httpClient *resty.Client
	token      *Token
}

func NewMiniprogram(config *Config) *Miniprogram {
	if err := config.Check(); err != nil {
		panic(err)
	}
	httpClient := resty.New()
	httpClient.SetBaseURL("https://api.weixin.qq.com")
	return &Miniprogram{
		config:     config,
		httpClient: httpClient,
		token: &Token{
			mutex: &sync.Mutex{},
		},
	}
}

// Config 获取配置
func (wx *Miniprogram) Config() *Config {
	return wx.config
}

// Code2Session 小程序登录
func (wx *Miniprogram) Code2Session(code string) (*sns.Jscode2sessionResponse, error) {
	return sns.Jscode2session(wx.httpClient, wx.config.AppId, wx.config.Secret, code)
}

// GetAccessToken 获取接口调用凭据
func (wx *Miniprogram) GetAccessToken() (*cgibin.TokenResponse, error) {
	wx.token.mutex.Lock()
	defer wx.token.mutex.Unlock()

	if wx.token.NetTime > 0 && wx.token.ExpiresIn > 0 && wx.token.ExpiresIn > (time.Now().Unix()-wx.token.NetTime) {
		return &cgibin.TokenResponse{
			AccessToken: wx.token.AccessToken,
			ExpiresIn:   wx.token.ExpiresIn,
		}, nil
	}
	token, err := cgibin.Token(wx.httpClient, wx.config.AppId, wx.config.Secret)
	if err != nil {
		return nil, err
	}
	wx.token.AccessToken = token.AccessToken
	if token.ExpiresIn > 0 {
		if token.ExpiresIn >= 7200 {
			wx.token.ExpiresIn = token.ExpiresIn - 300
			wx.token.NetTime = time.Now().Unix()
		} else {
			wx.token.ExpiresIn = int64(float64(token.ExpiresIn) * 0.8)
			if wx.token.ExpiresIn < 300 {
				wx.token.ExpiresIn = 0
			} else {
				wx.token.NetTime = time.Now().Unix()
			}
		}
	} else {
		wx.token.ExpiresIn = 0
	}
	return token, nil
}

// GetPhoneNumber 获取手机号
func (wx *Miniprogram) GetPhoneNumber(code string) (*wxa.GetUserPhoneNumberResponse, error) {
	getAccessToken, err := wx.GetAccessToken()
	if err != nil {
		return nil, err
	}
	return wxa.GetUserPhoneNumber(wx.httpClient, getAccessToken.AccessToken, code)
}

// GenerateScheme 获取加密 scheme 码
func (wx *Miniprogram) GenerateScheme(req *wxa.GenerateSchemeRequest) (*wxa.GenerateSchemeResponse, error) {
	getAccessToken, err := wx.GetAccessToken()
	if err != nil {
		return nil, err
	}
	return wxa.GenerateScheme(wx.httpClient, getAccessToken.AccessToken, req)
}

// QueryScheme 查询 scheme 码
func (wx *Miniprogram) QueryScheme(req *wxa.QuerySchemeRequest) (*wxa.QuerySchemeResponse, error) {
	getAccessToken, err := wx.GetAccessToken()
	if err != nil {
		return nil, err
	}
	return wxa.QueryScheme(wx.httpClient, getAccessToken.AccessToken, req)
}
