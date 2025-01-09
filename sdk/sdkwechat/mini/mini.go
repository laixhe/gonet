package mini

import (
	"context"
	"errors"
	"fmt"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
	authResponse "github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram/auth/response"
	phoneNumberResponse "github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram/phoneNumber/response"

	"github.com/laixhe/gonet/protocol/gen/config/cwechat"
	"github.com/laixhe/gonet/xlog"
)

// 微信小程序

type SdkWeChatMiniProgram struct {
	c      *cwechat.MiniProgram
	client *miniProgram.MiniProgram
}

func (sdk *SdkWeChatMiniProgram) Client() *miniProgram.MiniProgram {
	return sdkMiniProgram.client
}

var sdkMiniProgram *SdkWeChatMiniProgram

func SDK() *SdkWeChatMiniProgram {
	return sdkMiniProgram
}

// Init 初始化小程序
func Init(c *cwechat.MiniProgram) error {
	if c == nil {
		return errors.New("wechat mini program config as nil")
	}
	if c.AppId == "" {
		return errors.New("wechat mini program config appid as empty")
	}
	if c.Secret == "" {
		return errors.New("wechat mini program config secret as empty")
	}
	if c.Token == "" {
		//return errors.New("wechat mini program config token as empty")
	}
	if c.Aeskey == "" {
		//return errors.New("wechat mini program config aeskey as empty")
	}
	xlog.Debugf("wechat mini program config=%v", c)
	// doc https://powerwechat.artisan-cloud.com/zh/mini-program
	client, err := miniProgram.NewMiniProgram(&miniProgram.UserConfig{
		AppID:     c.AppId,
		Secret:    c.Secret,
		Token:     c.Token,
		AESKey:    c.Aeskey,
		HttpDebug: true,
		Debug:     true,
		Log:       miniProgram.Log{Stdout: true},
	})
	if err != nil {
		return err
	}
	//
	sdkMiniProgram = &SdkWeChatMiniProgram{
		c:      c,
		client: client,
	}
	return nil
}

// AuthSession 小程序登录
func AuthSession(ctx context.Context, code string) (*authResponse.ResponseCode2Session, error) {
	resp, err := sdkMiniProgram.client.Auth.Session(ctx, code)
	if err != nil {
		return nil, err
	}
	if resp.ErrCode > 0 {
		return nil, fmt.Errorf("%d %s", resp.ErrCode, resp.ErrMsg)
	}
	return resp, nil
}

// GetUserPhoneNumber 获取用户手机号
func GetUserPhoneNumber(ctx context.Context, code string) (*phoneNumberResponse.PhoneInfo, error) {
	resp, err := sdkMiniProgram.client.PhoneNumber.GetUserPhoneNumber(ctx, code)
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
