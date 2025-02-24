package wechatpay

// 交易类型
const (
	TRADE_TYPE_JSAPI    = "JSAPI"    // 公众号支付
	TRADE_TYPE_NATIVE   = "NATIVE"   // 扫码支付
	TRADE_TYPE_APP      = "APP"      // App支付
	TRADE_TYPE_MICROPAY = "MICROPAY" // 付款码支付
	TRADE_TYPE_MWEB     = "MWEB"     // H5支付
	TRADE_TYPE_FACEPAY  = "FACEPAY"  // 刷脸支付
)

// 交易状态
const (
	TRADE_STATE_SUCCESS    = "SUCCESS"    // 支付成功
	TRADE_STATE_REFUND     = "REFUND"     // 转入退款
	TRADE_STATE_NOTPAY     = "NOTPAY"     // 未支付
	TRADE_STATE_CLOSED     = "CLOSED"     // 已关闭
	TRADE_STATE_REVOKED    = "REVOKED"    // 已撤销（付款码支付）
	TRADE_STATE_USERPAYING = "USERPAYING" // 用户支付中（付款码支付）
	TRADE_STATE_PAYERROR   = "PAYERROR"   // 支付失败(其他原因，如银行返回失败)
)
