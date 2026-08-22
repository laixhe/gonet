package sms

import (
	"strings"
	"testing"

	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
)

// TestCheckSendSmsResp 覆盖 nil body、成功、Code 为 nil、业务失败四类场景
func TestCheckSendSmsResp(t *testing.T) {
	// nil body
	if err := checkSendSmsResp(nil); err == nil {
		t.Fatal("expected error for nil body")
	}
	// 成功
	okBody := &dysmsapi.SendSmsResponseBody{}
	okBody.SetCode("OK")
	if err := checkSendSmsResp(okBody); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	// Code 为 nil(响应体畸形)视为失败
	nilCode := &dysmsapi.SendSmsResponseBody{}
	if err := checkSendSmsResp(nilCode); err == nil {
		t.Fatal("expected error for nil code")
	}
	// 业务失败带消息
	failBody := &dysmsapi.SendSmsResponseBody{}
	failBody.SetCode("isv.SMS_SIGNATURE_ILLEGAL").SetMessage("签名不合法")
	err := checkSendSmsResp(failBody)
	if err == nil {
		t.Fatal("expected error for business failure")
	}
	if !strings.Contains(err.Error(), "签名不合法") {
		t.Fatalf("unexpected err: %v", err)
	}
}
