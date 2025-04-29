package mini

import (
	"context"
	"errors"
	"fmt"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
	authResponse "github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram/auth/response"
	phoneNumberResponse "github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram/phoneNumber/response"

	"github.com/laixhe/gonet/protocol/gen/config/cwechat"
	"github.com/laixhe/gonet/xlog"
)

// SdkWeChatMiniProgram 微信小程序
type SdkWeChatMiniProgram struct {
	config     *cwechat.MiniProgram     // 小程序配置
	client     *miniProgram.MiniProgram // 小程序客户端
	baseClient *kernel.BaseClient       // 小程序基础客户端
}

func (s *SdkWeChatMiniProgram) Config() *cwechat.MiniProgram {
	return s.config
}

func (s *SdkWeChatMiniProgram) Client() *miniProgram.MiniProgram {
	return s.client
}

func (s *SdkWeChatMiniProgram) BaseClient() *kernel.BaseClient {
	return s.baseClient
}

// AuthSession 小程序登录
func (s *SdkWeChatMiniProgram) AuthSession(ctx context.Context, code string) (*authResponse.ResponseCode2Session, error) {
	resp, err := s.client.Auth.Session(ctx, code)
	if err != nil {
		return nil, err
	}
	if resp.ErrCode > 0 {
		return nil, fmt.Errorf("%d %s", resp.ErrCode, resp.ErrMsg)
	}
	return resp, nil
}

// GetUserPhoneNumber 获取用户手机号
func (s *SdkWeChatMiniProgram) GetUserPhoneNumber(ctx context.Context, code string) (*phoneNumberResponse.PhoneInfo, error) {
	resp, err := s.client.PhoneNumber.GetUserPhoneNumber(ctx, code)
	if err != nil {
		return nil, err
	}
	if resp.ErrCode > 0 {
		return nil, fmt.Errorf("%d %s", resp.ErrCode, resp.ErrMsg)
	}
	if resp.PhoneInfo == nil {
		return nil, fmt.Errorf("phone info is nil; %d %s", resp.ErrCode, resp.ErrMsg)
	}
	return resp.PhoneInfo, nil
}

// Init 初始化小程序
func Init(config *cwechat.MiniProgram, isDebug bool) (*SdkWeChatMiniProgram, error) {
	if config == nil {
		return nil, errors.New("wechat mini program config as nil")
	}
	if config.AppId == "" {
		return nil, errors.New("wechat mini program config appid as empty")
	}
	if config.Secret == "" {
		return nil, errors.New("wechat mini program config secret as empty")
	}
	if config.Token == "" {
		//return nil, errors.New("wechat mini program config token as empty")
	}
	if config.Aeskey == "" {
		//return nil, errors.New("wechat mini program config aeskey as empty")
	}
	xlog.Debugf("wechat mini program config=%v", config)
	// doc https://powerwechat.artisan-cloud.com/zh/mini-program
	client, err := miniProgram.NewMiniProgram(&miniProgram.UserConfig{
		AppID:     config.AppId,
		Secret:    config.Secret,
		Token:     config.Token,
		AESKey:    config.Aeskey,
		HttpDebug: isDebug,
		Debug:     isDebug,
		Log:       miniProgram.Log{Stdout: true},
	})
	if err != nil {
		return nil, err
	}
	// 基础客户端
	baseClient, err := kernel.NewBaseClient(client, nil)
	if err != nil {
		return nil, err
	}
	return &SdkWeChatMiniProgram{
		config:     config,
		client:     client,
		baseClient: baseClient,
	}, nil
}
