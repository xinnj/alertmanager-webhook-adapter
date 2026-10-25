package weixin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestBot(t *testing.T, key string, status int, response string) *WeixinGroupBot {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if response != "" {
			_, _ = w.Write([]byte(response))
		}
	}))
	t.Cleanup(srv.Close)

	bot := NewWexinGroupBot(key)
	bot.addr = srv.URL
	return bot
}

func TestSendSucceedsOnZeroErrcode(t *testing.T) {
	bot := newTestBot(t, "my-key", http.StatusOK, `{"errcode":0,"errmsg":"ok"}`)
	if err := bot.Send(&Msg{}); err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}
}

func TestSendReturnsErrorOnBusinessError(t *testing.T) {
	bot := newTestBot(t, "my-key", http.StatusOK, `{"errcode":93000,"errmsg":"invalid webhook url"}`)

	err := bot.Send(&Msg{})
	if err == nil {
		t.Fatal("Send() error = nil, want error")
	}
	for _, want := range []string{"93000", "invalid webhook url"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err, want)
		}
	}
	if strings.Contains(err.Error(), "my-key") {
		t.Errorf("error %q leaks the webhook key", err)
	}
}

func TestSendReturnsErrorOnNonJSONBody(t *testing.T) {
	bot := newTestBot(t, "my-key", http.StatusOK, `<html>bad gateway</html>`)

	err := bot.Send(&Msg{})
	if err == nil {
		t.Fatal("Send() error = nil, want error")
	}
	if strings.Contains(err.Error(), "my-key") {
		t.Errorf("error %q leaks the webhook key", err)
	}
}

func TestSendErrorDoesNotLeakToken(t *testing.T) {
	bot := NewWexinGroupBot("my-key")
	bot.addr = "http://127.0.0.1:1"

	err := bot.Send(&Msg{})
	if err == nil {
		t.Fatal("Send() error = nil, want error")
	}
	if strings.Contains(err.Error(), "my-key") {
		t.Errorf("error %q leaks the webhook key", err)
	}
}
