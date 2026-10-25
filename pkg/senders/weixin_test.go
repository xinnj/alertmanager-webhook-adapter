package senders

import (
	"net/http"
	"net/http/httptest"
	"testing"

	restful "github.com/emicklei/go-restful/v3"
)

func newWeixinRequest(t *testing.T) *restful.Request {
	t.Helper()
	return restful.NewRequest(httptest.NewRequest(http.MethodPost, "/webhook/send?channel_type=weixin", nil))
}

func TestCreateWeixinSenderUsesEnvToken(t *testing.T) {
	t.Setenv("WEIXIN_TOKEN", "env-key")

	if _, err := createWeixinSender(newWeixinRequest(t)); err != nil {
		t.Fatalf("createWeixinSender() error = %v, want nil", err)
	}
}

func TestCreateWeixinSenderRequiresToken(t *testing.T) {
	t.Setenv("WEIXIN_TOKEN", "")

	if _, err := createWeixinSender(newWeixinRequest(t)); err == nil {
		t.Fatal("createWeixinSender() error = nil, want error")
	}
}
