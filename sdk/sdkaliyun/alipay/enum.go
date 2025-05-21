package alipay

// 通知类型
const (
	NOTIFY_TYPE_TRADE_STATUS_SYNC = "trade_status_sync" // 扣款通知(交易状态变更通知)
	NOTIFY_TYPE_DUT_USER_SIGN     = "dut_user_sign"     // 签约通知
	NOTIFY_TYPE_DUT_USER_UNSIGN   = "dut_user_unsign"   // 解约通知
	NOTIFY_TYPE_REFUND_FAILED     = "refund_failed"     // 拒绝退款
)

// 状态说明
const (
	TRADE_STATUS_WAIT_BUYER_PAY = "WAIT_BUYER_PAY" // 交易创建，等待买家付款
	TRADE_STATUS_CLOSED         = "TRADE_CLOSED"   // 未付款交易超时关闭，或支付完成后全额退款
	TRADE_STATUS_SUCCESS        = "TRADE_SUCCESS"  // 交易支付成功
	TRADE_STATUS_FINISHED       = "TRADE_FINISHED" // 交易结束，不可退款
)
