package wechatpay

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/go-pay/gopay"
	wechatV3 "github.com/go-pay/gopay/wechat/v3"
	"github.com/laixhe/gonet/xlog"
	"go.uber.org/zap"

	"github.com/laixhe/gonet/protocol/gen/config/cwechat"
	xginConstant "github.com/laixhe/gonet/xgin/constant"
)

type SdkWechatPay struct {
	c      *cwechat.WeChatPay
	client *wechatV3.ClientV3
}

var sdkWechatPay *SdkWechatPay

func Client() *wechatV3.ClientV3 {
	return sdkWechatPay.client
}

func Config() *cwechat.WeChatPay {
	return sdkWechatPay.c
}

func Init(c *cwechat.WeChatPay, isDebugLog bool) error {
	if c == nil {
		return errors.New("wechat pay config is nil")
	}
	if c.AppId == "" {
		panic("wechat pay config app_id is empty")
	}
	if c.Secret == "" {
		panic("wechat pay config secret is empty")
	}
	if c.MchId == "" {
		return errors.New("wechat pay config mch_id is empty")
	}
	if c.MchSerialNo == "" {
		return errors.New("wechat pay config mch_serial_no is empty")
	}
	if c.MchApiV3Key == "" {
		return errors.New("wechat pay config mch_api_v3_key is empty")
	}
	if c.MchPrivateKeyPath == "" {
		return errors.New("wechat pay config mch_private_key_path is empty")
	}
	if c.NotifyUrl == "" {
		return errors.New("wechat pay config notify_url is empty")
	}
	xlog.Debugf("wechat pay config=%v", c)
	//
	var err error
	var privateKey []byte
	privateKey, err = os.ReadFile(c.MchPrivateKeyPath)
	if err != nil {
		return errors.New("wechat pay private key error: " + err.Error())
	}
	client, err := wechatV3.NewClientV3(c.MchId, c.MchSerialNo, c.MchApiV3Key, string(privateKey))
	if err != nil {
		return errors.New("wechat pay new client error: " + err.Error())
	}
	// 启用自动同步返回验签，并定时更新微信平台API证书
	err = client.AutoVerifySign()
	if err != nil {
		return errors.New("wechat pay auto verify sign error: " + err.Error())
	}
	// 打开 Debug 开关，输出日志，默认是关闭的
	if isDebugLog {
		client.DebugSwitch = gopay.DebugOn
	}
	//
	sdkWechatPay = &SdkWechatPay{
		c:      c,
		client: client,
	}
	return nil
}

// AppPay APP支付(预支付交易会话标识)
// requestID   请求唯一值
// title       订单标题
// orderNumber 订单号
// money       订单总金额，单位为分
func AppPay(ctx context.Context, requestID string, title, orderNumber string, money uint64) (string, error) {
	bm := make(gopay.BodyMap)
	bm.Set("appid", sdkWechatPay.c.AppId)
	bm.Set("mchid", sdkWechatPay.c.MchId)
	bm.Set("description", title)
	bm.Set("out_trade_no", orderNumber)
	bm.Set("notify_url", sdkWechatPay.c.NotifyUrl)
	bm.SetBodyMap("amount", func(bm gopay.BodyMap) {
		bm.Set("total", money)
		bm.Set("currency", "CNY")
	})
	//
	resp, err := sdkWechatPay.client.V3TransactionApp(ctx, bm)
	if err != nil {
		xlog.Error(err.Error(), zap.String(xginConstant.HeaderRequestID, requestID))
		return "", err
	}
	if resp.Code != wechatV3.Success {
		return "", errors.New(resp.Error)
	}
	return resp.Response.PrepayId, nil
}

// PayNotify 支付异步回调
func PayNotify(req *http.Request, requestID string) (*wechatV3.V3DecryptPayResult, error) {
	// 解析
	notifyReq, err := wechatV3.V3ParseNotify(req)
	if err != nil {
		xlog.Error(err.Error(), zap.String(xginConstant.HeaderRequestID, requestID))
		return nil, err
	}
	// 获取微信平台证书
	certMap := sdkWechatPay.client.WxPublicKeyMap()
	// 验证异步通知的签名
	err = notifyReq.VerifySignByPKMap(certMap)
	if err != nil {
		xlog.Error(err.Error(), zap.String(xginConstant.HeaderRequestID, requestID))
		return nil, err
	}
	// 支付通知解密
	result, err := notifyReq.DecryptPayCipherText(sdkWechatPay.c.MchApiV3Key)
	if err != nil {
		xlog.Error(err.Error(), zap.String(xginConstant.HeaderRequestID, requestID))
		return nil, err
	}
	return result, nil
}
