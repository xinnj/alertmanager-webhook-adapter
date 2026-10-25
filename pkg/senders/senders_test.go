package senders

import (
	"net/http/httptest"
	"testing"

	restful "github.com/emicklei/go-restful/v3"
)

func newRequest(rawQuery string) *restful.Request {
	return restful.NewRequest(httptest.NewRequest("POST", "/webhook/send/?"+rawQuery, nil))
}

func TestCredentialFromQueryParameter(t *testing.T) {
	t.Setenv("DINGTALK_TOKEN", "env-token")

	got := credentialFrom(newRequest("token=query-token"), "token", "DINGTALK_TOKEN")
	if got != "query-token" {
		t.Errorf("credentialFrom() = %q, want %q", got, "query-token")
	}
}

func TestCredentialFromEnvironment(t *testing.T) {
	t.Setenv("DINGTALK_TOKEN", "env-token")

	got := credentialFrom(newRequest(""), "token", "DINGTALK_TOKEN")
	if got != "env-token" {
		t.Errorf("credentialFrom() = %q, want %q", got, "env-token")
	}
}

func TestDingtalkSenderFallsBackToEnv(t *testing.T) {
	t.Setenv("DINGTALK_TOKEN", "env-token")
	t.Setenv("DINGTALK_SECRET", "env-secret")

	if _, err := createDingtalkSender(newRequest("")); err != nil {
		t.Errorf("createDingtalkSender() error = %v", err)
	}
}

func TestDingtalkSenderRequiresToken(t *testing.T) {
	t.Setenv("DINGTALK_TOKEN", "")

	if _, err := createDingtalkSender(newRequest("")); err == nil {
		t.Error("createDingtalkSender() error = nil, want error")
	}
}

func TestFeishuSenderFallsBackToEnv(t *testing.T) {
	t.Setenv("FEISHU_TOKEN", "env-token")
	t.Setenv("FEISHU_SECRET", "env-secret")

	if _, err := createFeishuSender(newRequest("")); err != nil {
		t.Errorf("createFeishuSender() error = %v", err)
	}
}

func TestFeishuSenderRequiresToken(t *testing.T) {
	t.Setenv("FEISHU_TOKEN", "")

	if _, err := createFeishuSender(newRequest("")); err == nil {
		t.Error("createFeishuSender() error = nil, want error")
	}
}
