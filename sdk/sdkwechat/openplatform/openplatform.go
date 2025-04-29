package openplatform

import (
	"errors"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/openPlatform"
	"github.com/laixhe/gonet/protocol/gen/config/cwechat"
)

// SdkWeChatOpenProgram 微信开放平台
type SdkWeChatOpenProgram struct {
	config     *cwechat.OpenProgram       // 开放平台配置
	client     *openPlatform.OpenPlatform // 开放平台客户端
	baseClient *kernel.BaseClient         // 开放平台基础客户端
}

func (s *SdkWeChatOpenProgram) Config() *cwechat.OpenProgram {
	return s.config
}

func (s *SdkWeChatOpenProgram) Client() *openPlatform.OpenPlatform {
	return s.client
}

func (s *SdkWeChatOpenProgram) BaseClient() *kernel.BaseClient {
	return s.baseClient
}

// Init 初始化开放平台
func Init(config *cwechat.OpenProgram, isDebug bool) (*SdkWeChatOpenProgram, error) {
	if config == nil {
		return nil, errors.New("wechat open program config as nil")
	}
	if config.AppId == "" {
		return nil, errors.New("wechat open program config appid as empty")
	}
	if config.Secret == "" {
		return nil, errors.New("wechat open program config secret as empty")
	}
	if config.Token == "" {
		//return nil, errors.New("wechat open program config token as empty")
	}
	if config.Aeskey == "" {
		//return nil, errors.New("wechat open program config aeskey as empty")
	}
	client, err := openPlatform.NewOpenPlatform(&openPlatform.UserConfig{
		AppID:     config.AppId,
		Secret:    config.Secret,
		Token:     config.Token,
		AESKey:    config.Aeskey,
		HttpDebug: isDebug,
		Debug:     isDebug,
		Log:       openPlatform.Log{Stdout: true},
	})
	if err != nil {
		return nil, err
	}
	// 基础客户端
	baseClient, err := kernel.NewBaseClient(client, nil)
	if err != nil {
		return nil, err
	}
	return &SdkWeChatOpenProgram{
		config:     config,
		client:     client,
		baseClient: baseClient,
	}, nil
}
