package douyin

import (
	"errors"
	"testing"

	openApiSdkClient "github.com/bytedance/douyin-openapi-sdk-go/client"
)

// newJsCode2SessionResp 构造 code2session 响应
func newJsCode2SessionResp(data *openApiSdkClient.V2Jscode2sessionResponseData, errNo int64, errTips string) *openApiSdkClient.V2Jscode2sessionResponse {
	resp := &openApiSdkClient.V2Jscode2sessionResponse{}
	resp.SetErrNo(errNo)
	resp.SetErrTips(errTips)
	if data != nil {
		resp.SetData(data)
	}
	return resp
}

// TestParseJsCode2SessionUnionIDEmpty 验证 unionid 为空(未在开发者后台绑定)时登录不应失败
func TestParseJsCode2SessionUnionIDEmpty(t *testing.T) {
	data := &openApiSdkClient.V2Jscode2sessionResponseData{}
	data.SetOpenid("openid1")
	data.SetSessionKey("sk")
	got, err := parseJsCode2Session(newJsCode2SessionResp(data, 0, ""), "code1", "")
	if err != nil {
		t.Fatalf("unionid empty should not fail: %v", err)
	}
	if got.OpenID != "openid1" || got.UnionID != "" || got.SessionKey != "sk" || got.AnonymousOpenID != "" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

// TestParseJsCode2SessionUnionIDPresent 验证 unionid 正常返回的场景
func TestParseJsCode2SessionUnionIDPresent(t *testing.T) {
	data := &openApiSdkClient.V2Jscode2sessionResponseData{}
	data.SetOpenid("openid1")
	data.SetUnionid("union1")
	data.SetSessionKey("sk")
	got, err := parseJsCode2Session(newJsCode2SessionResp(data, 0, ""), "code1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.OpenID != "openid1" || got.UnionID != "union1" || got.SessionKey != "sk" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

// TestParseJsCode2SessionMissingOpenid 验证 openid 缺失时报业务错误
func TestParseJsCode2SessionMissingOpenid(t *testing.T) {
	data := &openApiSdkClient.V2Jscode2sessionResponseData{}
	data.SetUnionid("union1")
	_, err := parseJsCode2Session(newJsCode2SessionResp(data, 0, ""), "code1", "")
	if err == nil {
		t.Fatal("expected error when openid missing")
	}
	if !errors.Is(err, ErrBusiness) {
		t.Fatalf("expected ErrBusiness, got %v", err)
	}
}

// TestParseJsCode2SessionErrNo 验证请求级 err_no 非 0 优先返回
func TestParseJsCode2SessionErrNo(t *testing.T) {
	_, err := parseJsCode2Session(newJsCode2SessionResp(nil, 40001, "bad code"), "code1", "")
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if apiErr.ErrCode != 40001 || apiErr.ErrMsg != "bad code" {
		t.Fatalf("unexpected err: %+v", apiErr)
	}
	if !errors.Is(err, ErrBusiness) {
		t.Fatalf("expected ErrBusiness, got %v", err)
	}
}

// TestParseJsCode2SessionAnonymous 验证匿名登录正常场景
func TestParseJsCode2SessionAnonymous(t *testing.T) {
	data := &openApiSdkClient.V2Jscode2sessionResponseData{}
	data.SetAnonymousOpenid("anon1")
	got, err := parseJsCode2Session(newJsCode2SessionResp(data, 0, ""), "", "anon-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AnonymousOpenID != "anon1" || got.OpenID != "" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

// TestParseJsCode2SessionAnonymousMissing 验证匿名模式下 anonymous_openid 缺失时报业务错误
func TestParseJsCode2SessionAnonymousMissing(t *testing.T) {
	_, err := parseJsCode2Session(newJsCode2SessionResp(nil, 0, ""), "", "anon-code")
	if err == nil {
		t.Fatal("expected error when anonymous_openid missing")
	}
	if !errors.Is(err, ErrBusiness) {
		t.Fatalf("expected ErrBusiness, got %v", err)
	}
}

// TestParseJsCode2SessionNilResp 验证空响应报本地错误
func TestParseJsCode2SessionNilResp(t *testing.T) {
	_, err := parseJsCode2Session(nil, "code1", "")
	if err == nil {
		t.Fatal("expected error for nil resp")
	}
	if !errors.Is(err, ErrLocal) {
		t.Fatalf("expected ErrLocal, got %v", err)
	}
}

// TestParseJsCode2SessionFallbackErr 验证 err_no/err_tips 缺失时回退 ECodeCall/通用提示
func TestParseJsCode2SessionFallbackErr(t *testing.T) {
	// data 无 openid 且 err_no/err_tips 均缺失
	resp := &openApiSdkClient.V2Jscode2sessionResponse{}
	resp.SetData(&openApiSdkClient.V2Jscode2sessionResponseData{})
	_, err := parseJsCode2Session(resp, "code1", "")
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if apiErr.ErrCode != ECodeCall || apiErr.ErrMsg != "调用失败" {
		t.Fatalf("unexpected err: %+v", apiErr)
	}
}

// TestMiniLoginJsCode2SessionBothEmpty 验证 code 与 anonymous_code 同时为空时本地拦截
func TestMiniLoginJsCode2SessionBothEmpty(t *testing.T) {
	d := &Douyin{config: &Config{AppID: "appid", AppSecret: "secret"}}
	_, err := d.MiniLoginJsCode2Session("")
	if err == nil {
		t.Fatal("expected error when both code and anonymous_code empty")
	}
	if !errors.Is(err, ErrLocal) {
		t.Fatalf("expected ErrLocal, got %v", err)
	}
}
