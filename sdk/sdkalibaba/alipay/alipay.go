package alipay

import (
	"context"
	"errors"
	"os"

	"github.com/go-pay/gopay"
	alipayV3 "github.com/go-pay/gopay/alipay/v3"
	"github.com/laixhe/gonet/xlog"

	"github.com/laixhe/gonet/protocol/gen/config/calipay"
)

type SdkAlipay struct {
	c                        *calipay.Alipay
	client                   *alipayV3.ClientV3
	AppCertPublicKeyBytes    []byte // 应用公钥证书
	AlipayRootCertBytes      []byte // 支付宝根证书
	AlipayCertPublicKeyBytes []byte // 支付宝公钥证书
}

var sdkAlipay *SdkAlipay

func Client() *alipayV3.ClientV3 {
	return sdkAlipay.client
}

func Config() *calipay.Alipay {
	return sdkAlipay.c
}

func Init(c *calipay.Alipay, isDebugLog bool) error {
	if c == nil {
		return errors.New("alipay config is nil")
	}
	xlog.Debugf("alipay config=%v", c)
	//
	var appCertPublicKeyBytes []byte    // 应用公钥证书
	var alipayRootCertBytes []byte      // 支付宝根证书
	var alipayCertPublicKeyBytes []byte // 支付宝公钥证书
	var err error
	//
	if c.AppId == "" {
		panic("alipay config appId is empty")
	}
	if c.PrivateKey == "" {
		panic("alipay config privateKey is empty")
	}

	if c.AppCertPublicKeyFile == "" {
		return errors.New("alipay config app_cert_public_key_file is empty")
	}
	if c.AlipayRootCertFile == "" {
		return errors.New("alipay config alipay_root_cert_file is empty")
	}
	if c.AlipayCertPublicKeyFile == "" {
		return errors.New("alipay config alipay_cert_public_key_file is empty")
	}

	appCertPublicKeyBytes, err = os.ReadFile(c.AppCertPublicKeyFile)
	if err != nil {
		return errors.New("alipay config app_cert_public_key_file error: " + err.Error())
	}
	alipayRootCertBytes, err = os.ReadFile(c.AlipayRootCertFile)
	if err != nil {
		return errors.New("alipay config alipay_root_cert_file error: " + err.Error())
	}
	alipayCertPublicKeyBytes, err = os.ReadFile(c.AlipayCertPublicKeyFile)
	if err != nil {
		return errors.New("alipay config alipay_cert_public_key_file error: " + err.Error())
	}
	// 初始化支付宝客V3户端
	client, err := alipayV3.NewClientV3(c.AppId, c.PrivateKey, c.IsProduction)
	if err != nil {
		return errors.New("alipay new client error: " + err.Error())
	}
	// 打开 Debug 开关，输出日志，默认是关闭的
	if isDebugLog {
		client.DebugSwitch = gopay.DebugOn
	}
	// 传入证书内容
	err = client.SetCert(appCertPublicKeyBytes, alipayRootCertBytes, alipayCertPublicKeyBytes)
	if err != nil {
		return errors.New("alipay set cert error: " + err.Error())
	}
	//
	sdkAlipay = &SdkAlipay{
		c:                        c,
		client:                   client,
		AppCertPublicKeyBytes:    appCertPublicKeyBytes,
		AlipayRootCertBytes:      alipayRootCertBytes,
		AlipayCertPublicKeyBytes: alipayCertPublicKeyBytes,
	}
	return nil
}

// AppPay APP支付(预支付交易会话标识)
// title       订单标题
// orderNumber 订单号
// money       订单总金额，单位为元，精确到小数点后两位 10.00
func AppPay(ctx context.Context, title, orderNumber, money string) (string, error) {
	bm := make(gopay.BodyMap)
	bm.Set("subject", title)
	bm.Set("out_trade_no", orderNumber)
	bm.Set("total_amount", money)
	bm.Set("notify_url", sdkAlipay.c.NotifyUrl)
	return "", nil
}
