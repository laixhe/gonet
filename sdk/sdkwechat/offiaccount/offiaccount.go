package offiaccount

import (
	"errors"

	"github.com/ArtisanCloud/PowerSocialite/v3/src/providers"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"

	"github.com/laixhe/gonet/protocol/gen/config/cwechat"
	"github.com/laixhe/gonet/xlog"
)

// SdkWeChatOffiaccount 微信公众号
type SdkWeChatOffiaccount struct {
	config     *cwechat.Offiaccount
	client     *officialAccount.OfficialAccount
	baseClient *kernel.BaseClient
}

func (s *SdkWeChatOffiaccount) Config() *cwechat.Offiaccount {
	return s.config
}

func (s *SdkWeChatOffiaccount) Client() *officialAccount.OfficialAccount {
	return s.client
}

func (s *SdkWeChatOffiaccount) BaseClient() *kernel.BaseClient {
	return s.baseClient
}

// AccessToken APP微信登录(通过 code 获取用户 access_token)
func (s *SdkWeChatOffiaccount) AccessToken(code string) (*providers.User, error) {
	return s.client.OAuth.UserFromCode(code)
}

// UserInfo 获取用户基本信息
func (s *SdkWeChatOffiaccount) UserInfo(openId, accessToken string) (*providers.User, error) {
	return s.client.OAuth.UserFromToken(accessToken, openId)
}

// Init 初始化公众号
func Init(config *cwechat.Offiaccount, isDebug bool) (*SdkWeChatOffiaccount, error) {
	if config == nil {
		return nil, errors.New("wechat offiaccount config as nil")
	}
	if config.AppId == "" {
		return nil, errors.New("wechat offiaccount config appid as empty")
	}
	if config.Secret == "" {
		return nil, errors.New("wechat offiaccount config secret as empty")
	}
	xlog.Debugf("wechat offiaccount config=%v", config)
	// doc https://powerwechat.artisan-cloud.com/zh/official-account
	client, err := officialAccount.NewOfficialAccount(&officialAccount.UserConfig{
		AppID:  config.AppId,
		Secret: config.Secret,
		OAuth: officialAccount.OAuth{
			Scopes: []string{"snsapi_userinfo"},
		},
		HttpDebug: isDebug,
		Debug:     isDebug,
		Log:       officialAccount.Log{Stdout: true},
	})
	if err != nil {
		return nil, err
	}
	// 基础客户端
	baseClient, err := kernel.NewBaseClient(client, nil)
	if err != nil {
		return nil, err
	}
	return &SdkWeChatOffiaccount{
		config:     config,
		client:     client,
		baseClient: baseClient,
	}, nil
}
