package work

import (
	"sync"
	"time"

	"resty.dev/v3"

	"github.com/laixhe/gonet/wechat/work/cgibin"
)

type Token struct {
	mutex       *sync.Mutex
	NetTime     int64  // 最新时间戳
	ExpiresIn   int64  // 凭证有效时间，单位：秒。目前是 7200 秒之内的值(2个小时)
	AccessToken string // 获取到的凭证
}

// Work 企业微信
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
	return &Work{
		config:     config,
		httpClient: httpClient,
	}
}

// Config 获取配置
func (w *Work) Config() *Config {
	return w.config
}

// GetToken 获取接口调用凭据
func (w *Work) GetToken() (*cgibin.GetTokenResponse, error) {
	w.token.mutex.Lock()
	defer w.token.mutex.Unlock()

	if w.token.NetTime > 0 && w.token.ExpiresIn > 0 && w.token.ExpiresIn > (time.Now().Unix()-w.token.NetTime) {
		return &cgibin.GetTokenResponse{
			AccessToken: w.token.AccessToken,
			ExpiresIn:   w.token.ExpiresIn,
		}, nil
	}
	tokenResp, err := cgibin.GetToken(w.httpClient, w.config.Corpid, w.config.Corpid)
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
	return tokenResp, nil
}

// GetUserInfo 获取访问用户身份
func (w *Work) GetUserInfo(code string) (*cgibin.GetUserInfoResponse, error) {
	getAccessToken, err := w.GetToken()
	if err != nil {
		return nil, err
	}
	return cgibin.GetUserInfo(w.httpClient, getAccessToken.AccessToken, code)
}

// GetUserDetail 获取访问用户敏感信息
func (w *Work) GetUserDetail(userTicket string) (*cgibin.GetUserDetailResponse, error) {
	getAccessToken, err := w.GetToken()
	if err != nil {
		return nil, err
	}
	return cgibin.GetUserDetail(w.httpClient, getAccessToken.AccessToken, userTicket)
}

// UserGet 读取成员
func (w *Work) UserGet(userid string) (*cgibin.UserGetResponse, error) {
	getAccessToken, err := w.GetToken()
	if err != nil {
		return nil, err
	}
	return cgibin.UserGet(w.httpClient, getAccessToken.AccessToken, userid)
}
