package douyin

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/bytedance/douyin-openapi-credential-go/client"
	openApiSdkClient "github.com/bytedance/douyin-openapi-sdk-go/client"
)

// 抖音开放平台

// ECodeCall 调用失败码
const ECodeCall = -2

type Token struct {
	mutex       *sync.Mutex
	NetTime     int64  // 最新时间戳
	ExpiresIn   int64  // 凭证有效时间，单位：秒。目前是 7200 秒之内的值(2个小时)
	AccessToken string // 获取到的凭证
}

type Douyin struct {
	config *Config
	client *openApiSdkClient.Client
	token  *Token
}

func NewDouyin(config *Config) (*Douyin, error) {
	if err := config.Check(); err != nil {
		return nil, err
	}
	opt := new(credential.Config).
		SetClientKey(config.AppID).
		SetClientSecret(config.AppSecret)
	client, err := openApiSdkClient.NewClient(opt)
	if err != nil {
		return nil, err
	}
	return &Douyin{
		config: config,
		client: client,
		token: &Token{
			mutex: &sync.Mutex{},
		},
	}, nil
}

func (d *Douyin) Config() *Config {
	return d.config
}

func (d *Douyin) ClientToken() (string, error) {
	d.token.mutex.Lock()
	defer d.token.mutex.Unlock()

	if d.token.NetTime > 0 && d.token.ExpiresIn > 0 && d.token.ExpiresIn > (time.Now().Unix()-d.token.NetTime) {
		return d.token.AccessToken, nil
	}
	//
	req := &openApiSdkClient.OauthClientTokenRequest{}
	req.SetClientKey(d.config.AppID)
	req.SetClientSecret(d.config.AppSecret)
	req.SetGrantType("client_credential")
	resp, err := d.client.OauthClientToken(req)
	if err != nil {
		return "", err
	}
	message := tea.StringValue(resp.Message)
	if resp.Data == nil {
		if message == "" {
			message = "获取凭证失败"
		}
		return "", errors.New(message)
	}

	respErrorCode := tea.Int64Value(resp.Data.ErrorCode)
	respDescription := tea.StringValue(resp.Data.Description)
	respAccessToken := tea.StringValue(resp.Data.AccessToken)
	respExpiresIn := tea.Int64Value(resp.Data.ExpiresIn)

	if respErrorCode != 0 {
		if message == "" {
			message = "获取凭证失败 "
		}
		message += respDescription
		return "", errors.New(message)
	}
	if respAccessToken == "" {
		if message == "" {
			message = "获取凭证失败 "
		}
		message += respDescription
		return "", errors.New(message)
	}
	//
	d.token.AccessToken = respAccessToken
	d.token.NetTime = time.Now().Unix()
	if respExpiresIn > 200 {
		d.token.ExpiresIn = respExpiresIn - 200
	} else {
		d.token.ExpiresIn = 0
	}
	return respAccessToken, nil
}

// RsaDecryptByPrivateKeyStr 私钥解密
func (d *Douyin) RsaDecryptByPrivateKeyStr(cipherData string) (originText string, err error) {
	if d.config.PrivateKey == "" {
		return "", errors.New("没有私钥")
	}
	// 读取私钥
	privateKeyBytes, err := base64.StdEncoding.DecodeString(d.config.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("base64 解码私钥失败: %v", err)
	}
	// 解析私钥
	privateRSA, err := x509.ParsePKCS1PrivateKey(privateKeyBytes)
	if err != nil {
		return "", fmt.Errorf("解析私钥失败: %v", err)
	}
	// 读取数据
	cipherDataBytes, err := base64.StdEncoding.DecodeString(cipherData)
	if err != nil {
		return "", fmt.Errorf("base64 解码数据失败: %v", err)
	}
	// 解密数据
	originTextBytes, err := rsa.DecryptPKCS1v15(rand.Reader, privateRSA, cipherDataBytes)
	if err != nil {
		return "", fmt.Errorf("解密数据失败: %v", err)
	}
	return string(originTextBytes), nil
}
