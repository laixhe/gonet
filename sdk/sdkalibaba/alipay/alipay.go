package alipay

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/go-pay/gopay"
	alipayV2 "github.com/go-pay/gopay/alipay"
	"go.uber.org/zap"

	"github.com/laixhe/gonet/protocol/gen/config/calipay"
	xginConstant "github.com/laixhe/gonet/xgin/constant"
	"github.com/laixhe/gonet/xlog"
)

type SdkAlipay struct {
	config                   *calipay.Alipay
	client                   *alipayV2.Client
	AppCertPublicKeyBytes    []byte // 应用公钥证书
	AlipayRootCertBytes      []byte // 支付宝根证书
	AlipayCertPublicKeyBytes []byte // 支付宝公钥证书
}

func (s *SdkAlipay) Config() *calipay.Alipay {
	return s.config
}

func (s *SdkAlipay) Client() *alipayV2.Client {
	return s.client
}

// AppPay APP支付(预支付交易会话标识)
// requestID   请求唯一值
// title       订单标题
// orderNumber 订单号
// money       订单总金额，单位为元，精确到小数点后两位 10.00
func (s *SdkAlipay) AppPay(ctx context.Context, requestID string, title, orderNumber, money string) (string, error) {
	bm := make(gopay.BodyMap)
	bm.Set("subject", title)
	bm.Set("out_trade_no", orderNumber)
	bm.Set("total_amount", money)
	//bm.Set("notify_url", sdkAlipay.c.NotifyUrl)
	//
	resp, err := s.client.TradeAppPay(ctx, bm)
	if err != nil {
		xlog.Error(err.Error(), zap.String(xginConstant.HeaderRequestID, requestID))
		return "", err
	}
	return resp, nil
}

// PayNotify 支付异步回调
func (s *SdkAlipay) PayNotify(req *http.Request, requestID string) (map[string]any, error) {
	// 解析异步通知的参数
	notifyReq, err := alipayV2.ParseNotifyToBodyMap(req)
	if err != nil {
		xlog.Error(err.Error(), zap.String(xginConstant.HeaderRequestID, requestID))
		return nil, err
	}
	// 支付宝异步通知验签（公钥证书模式）
	_, err = alipayV2.VerifySignWithCert(s.AlipayCertPublicKeyBytes, notifyReq)
	if err != nil {
		xlog.Error(err.Error(), zap.String(xginConstant.HeaderRequestID, requestID))
		return nil, err
	}
	return notifyReq, nil
}

func Init(config *calipay.Alipay, isDebugLog bool) (*SdkAlipay, error) {
	if config == nil {
		return nil, errors.New("alipay config is nil")
	}
	xlog.Debugf("alipay config=%v", config)
	//
	var appCertPublicKeyBytes []byte    // 应用公钥证书
	var alipayRootCertBytes []byte      // 支付宝根证书
	var alipayCertPublicKeyBytes []byte // 支付宝公钥证书
	var err error
	//
	if config.AppId == "" {
		return nil, errors.New("alipay config appId is empty")
	}
	if config.PrivateKey == "" {
		return nil, errors.New("alipay config privateKey is empty")
	}
	if config.AppCertPublicKeyFile == "" {
		return nil, errors.New("alipay config app_cert_public_key_file is empty")
	}
	if config.AlipayRootCertFile == "" {
		return nil, errors.New("alipay config alipay_root_cert_file is empty")
	}
	if config.AlipayCertPublicKeyFile == "" {
		return nil, errors.New("alipay config alipay_cert_public_key_file is empty")
	}
	//
	appCertPublicKeyBytes, err = os.ReadFile(config.AppCertPublicKeyFile)
	if err != nil {
		return nil, errors.New("alipay config app_cert_public_key_file error: " + err.Error())
	}
	alipayRootCertBytes, err = os.ReadFile(config.AlipayRootCertFile)
	if err != nil {
		return nil, errors.New("alipay config alipay_root_cert_file error: " + err.Error())
	}
	alipayCertPublicKeyBytes, err = os.ReadFile(config.AlipayCertPublicKeyFile)
	if err != nil {
		return nil, errors.New("alipay config alipay_cert_public_key_file error: " + err.Error())
	}
	// 初始化支付宝客户端
	client, err := alipayV2.NewClient(config.AppId, config.PrivateKey, config.IsProduction)
	if err != nil {
		return nil, errors.New("alipay new client error: " + err.Error())
	}
	// 打开 Debug 开关，输出日志，默认是关闭的
	if isDebugLog {
		client.DebugSwitch = gopay.DebugOn
	}
	// 设置支付宝请求 公共参数
	client.SetReturnUrl(config.ReturnUrl) // 设置同步通知URL
	client.SetNotifyUrl(config.NotifyUrl) // 设置异步通知URL
	// 自动同步验签（只支持证书模式）
	// 传入 alipayPublicCert.crt 内容
	client.AutoVerifySign(alipayCertPublicKeyBytes)

	// 传入证书内容
	err = client.SetCertSnByContent(appCertPublicKeyBytes, alipayRootCertBytes, alipayCertPublicKeyBytes)
	if err != nil {
		return nil, errors.New("alipay set cert error: " + err.Error())
	}
	return &SdkAlipay{
		config:                   config,
		client:                   client,
		AppCertPublicKeyBytes:    appCertPublicKeyBytes,
		AlipayRootCertBytes:      alipayRootCertBytes,
		AlipayCertPublicKeyBytes: alipayCertPublicKeyBytes,
	}, nil
}
