package douyin

import (
	"errors"
	"fmt"
)

// 错误类别哨兵,配合 errors.Is 判断错误类型,例如:
//
//	if errors.Is(err, ErrNetwork) { ... }
var (
	// ErrLocal 本地错误(参数校验、响应为空、本地处理失败等)
	ErrLocal = errors.New("douyin api: local error")
	// ErrNetwork 网络错误(连接失败、超时、上下文取消)
	ErrNetwork = errors.New("douyin api: network error")
	// ErrHTTP HTTP 非 2xx 响应
	ErrHTTP = errors.New("douyin api: http error")
	// ErrBusiness 业务错误(err_no/code 非 0)
	ErrBusiness = errors.New("douyin api: business error")
	// ErrDecode 响应体解析失败
	ErrDecode = errors.New("douyin api: decode error")
)

// ErrKind 错误类别。
type ErrKind int

const (
	// ErrKindLocal 本地错误
	ErrKindLocal ErrKind = iota + 1
	// ErrKindNetwork 网络错误
	ErrKindNetwork
	// ErrKindHTTP HTTP 非 2xx 响应
	ErrKindHTTP
	// ErrKindBusiness 业务错误
	ErrKindBusiness
	// ErrKindDecode 响应体解析失败
	ErrKindDecode
)

// ApiError 抖音接口调用错误。
//
// 通过 errors.As 提取错误详情,通过 errors.Is 判断错误类别:
//
//	var apiErr *ApiError
//	if errors.As(err, &apiErr) { _ = apiErr.ErrCode }
//	if errors.Is(err, ErrNetwork) { ... }
type ApiError struct {
	Kind    ErrKind // 错误类别
	ErrCode int     // 错误码,业务错误时为抖音 err_no,本地失败为 ECodeCall
	ErrMsg  string  // 错误信息
	Status  int     // HTTP 状态码,仅 ErrKindHTTP 时有效
	kindErr error   // 对应的哨兵错误
}

// Unwrap 返回错误类别对应的哨兵错误,支持 errors.Is 判断。
func (e *ApiError) Unwrap() error {
	return e.kindErr
}

// newError 构造 ApiError 并关联哨兵错误。
func newError(kind ErrKind, errCode, status int, msg string) *ApiError {
	var sentinel error
	switch kind {
	case ErrKindLocal:
		sentinel = ErrLocal
	case ErrKindNetwork:
		sentinel = ErrNetwork
	case ErrKindHTTP:
		sentinel = ErrHTTP
	case ErrKindBusiness:
		sentinel = ErrBusiness
	case ErrKindDecode:
		sentinel = ErrDecode
	}
	return &ApiError{Kind: kind, ErrCode: errCode, ErrMsg: msg, Status: status, kindErr: sentinel}
}

func (e *ApiError) Error() string {
	switch e.Kind {
	case ErrKindHTTP:
		return fmt.Sprintf("douyin api error: http %d, %s", e.Status, e.ErrMsg)
	case ErrKindBusiness:
		return fmt.Sprintf("douyin api error: code %d, %s", e.ErrCode, e.ErrMsg)
	default:
		return fmt.Sprintf("douyin api error: %s", e.ErrMsg)
	}
}
