package wechatpay

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/go-pay/gopay"
	wechatV3 "github.com/go-pay/gopay/wechat/v3"
	"go.uber.org/zap"

	"github.com/laixhe/gonet/protocol/gen/config/cwechat"
	xginConstant "github.com/laixhe/gonet/xgin/constant"
	"github.com/laixhe/gonet/xlog"
)

type SdkWechatPay struct {
	config *cwechat.WeChatPay
	client *wechatV3.ClientV3
}

func (s *SdkWechatPay) Config() *cwechat.WeChatPay {
	return s.config
}

func (s *SdkWechatPay) Client() *wechatV3.ClientV3 {
	return s.client
}

// AppPay APP支付(预支付交易会话标识)
// requestID   请求唯一值
// title       订单标题
// orderNumber 订单号
// money       订单总金额，单位为分
func (s *SdkWechatPay) AppPay(ctx context.Context, requestID string, title, orderNumber string, money uint64) (string, error) {
	bm := make(gopay.BodyMap)
	bm.Set("appid", s.config.AppId)
	bm.Set("mchid", s.config.MchId)
	bm.Set("description", title)
	bm.Set("out_trade_no", orderNumber)
	bm.Set("notify_url", s.config.NotifyUrl)
	bm.SetBodyMap("amount", func(bm gopay.BodyMap) {
		bm.Set("total", money)
		bm.Set("currency", "CNY")
	})
	//
	resp, err := s.client.V3TransactionApp(ctx, bm)
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
func (s *SdkWechatPay) PayNotify(req *http.Request, requestID string) (*wechatV3.V3DecryptPayResult, error) {
	// 解析异步通知的参数
	notifyReq, err := wechatV3.V3ParseNotify(req)
	if err != nil {
		xlog.Error(err.Error(), zap.String(xginConstant.HeaderRequestID, requestID))
		return nil, err
	}
	// 获取微信平台证书
	certMap := s.client.WxPublicKeyMap()
	// 验证异步通知的签名
	err = notifyReq.VerifySignByPKMap(certMap)
	if err != nil {
		xlog.Error(err.Error(), zap.String(xginConstant.HeaderRequestID, requestID))
		return nil, err
	}
	// 支付通知解密
	result, err := notifyReq.DecryptPayCipherText(s.config.MchApiV3Key)
	if err != nil {
		xlog.Error(err.Error(), zap.String(xginConstant.HeaderRequestID, requestID))
		return nil, err
	}
	return result, nil
}

func Init(config *cwechat.WeChatPay, isDebugLog bool) (*SdkWechatPay, error) {
	if config == nil {
		return nil, errors.New("wechat pay config is nil")
	}
	if config.AppId == "" {
		return nil, errors.New("wechat pay config app_id is empty")
	}
	if config.Secret == "" {
		return nil, errors.New("wechat pay config secret is empty")
	}
	if config.MchId == "" {
		return nil, errors.New("wechat pay config mch_id is empty")
	}
	if config.MchSerialNo == "" {
		return nil, errors.New("wechat pay config mch_serial_no is empty")
	}
	if config.MchApiV3Key == "" {
		return nil, errors.New("wechat pay config mch_api_v3_key is empty")
	}
	if config.MchPrivateKeyPath == "" {
		return nil, errors.New("wechat pay config mch_private_key_path is empty")
	}
	if config.NotifyUrl == "" {
		return nil, errors.New("wechat pay config notify_url is empty")
	}
	xlog.Debugf("wechat pay config=%v", config)
	//
	var err error
	var privateKey []byte
	privateKey, err = os.ReadFile(config.MchPrivateKeyPath)
	if err != nil {
		return nil, errors.New("wechat pay private key error: " + err.Error())
	}
	client, err := wechatV3.NewClientV3(config.MchId, config.MchSerialNo, config.MchApiV3Key, string(privateKey))
	if err != nil {
		return nil, errors.New("wechat pay new client error: " + err.Error())
	}
	// 启用自动同步返回验签，并定时更新微信平台API证书
	err = client.AutoVerifySign()
	if err != nil {
		return nil, errors.New("wechat pay auto verify sign error: " + err.Error())
	}
	// 打开 Debug 开关，输出日志，默认是关闭的
	if isDebugLog {
		client.DebugSwitch = gopay.DebugOn
	}
	return &SdkWechatPay{
		config: config,
		client: client,
	}, nil
}
