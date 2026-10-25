package dingtalk

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func newTestDingtalkBot(t *testing.T, status int, response string) *DingtalkGroupBot {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if response != "" {
			_, _ = w.Write([]byte(response))
		}
	}))
	t.Cleanup(srv.Close)

	bot := NewDingtalkGroupBotWithSecret("my-token", "my-secret")
	bot.addr = srv.URL
	return bot
}

func TestSendSucceedsOnZeroErrcode(t *testing.T) {
	bot := newTestDingtalkBot(t, http.StatusOK, `{"errcode":0,"errmsg":"ok"}`)
	if err := bot.Send(&Msg{}); err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}
}

func TestSendReturnsErrorOnBusinessError(t *testing.T) {
	bot := newTestDingtalkBot(t, http.StatusOK, `{"errcode":310000,"errmsg":"invalid access token"}`)

	err := bot.Send(&Msg{})
	if err == nil {
		t.Fatal("Send() error = nil, want error")
	}
	for _, want := range []string{"310000", "invalid access token"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err, want)
		}
	}
	for _, credential := range []string{"my-token", "my-secret"} {
		if strings.Contains(err.Error(), credential) {
			t.Errorf("error %q contains %q", err, credential)
		}
	}
}

func TestSendReturnsErrorOnNonJSONBody(t *testing.T) {
	bot := newTestDingtalkBot(t, http.StatusOK, `<html>bad gateway</html>`)

	err := bot.Send(&Msg{})
	if err == nil {
		t.Fatal("Send() error = nil, want error")
	}
	if strings.Contains(err.Error(), "bad gateway") {
		t.Errorf("error %q unexpectedly includes the response body", err)
	}
}

func TestSendErrorDoesNotLeakCredentials(t *testing.T) {
	bot := NewDingtalkGroupBotWithSecret("my-token", "my-secret")
	bot.addr = "http://127.0.0.1:1"

	err := bot.Send(&Msg{})
	if err == nil {
		t.Fatal("Send() error = nil, want error")
	}

	for _, credential := range []string{"my-token", "my-secret", "sign="} {
		if strings.Contains(err.Error(), credential) {
			t.Errorf("error %q contains %q", err, credential)
		}
	}
}

func TestAddrWithoutSecretIsUnchanged(t *testing.T) {
	bot := NewDingtalkGroupBot("my-token")

	const want = "https://oapi.dingtalk.com/robot/send?access_token=my-token"
	if got := bot.Addr(); got != want {
		t.Errorf("Addr() = %q, want %q", got, want)
	}
}

func TestAddrWithSecretIsSigned(t *testing.T) {
	bot := NewDingtalkGroupBotWithSecret("my-token", "my-secret")

	u, err := url.Parse(bot.Addr())
	if err != nil {
		t.Fatalf("parse Addr(): %v", err)
	}

	q := u.Query()
	if got := q.Get("access_token"); got != "my-token" {
		t.Errorf("access_token = %q, want %q", got, "my-token")
	}

	ts, err := strconv.ParseInt(q.Get("timestamp"), 10, 64)
	if err != nil {
		t.Fatalf("timestamp %q is not an integer: %v", q.Get("timestamp"), err)
	}
	if delta := time.Since(time.UnixMilli(ts)); delta > time.Minute || delta < -time.Minute {
		t.Errorf("timestamp %v is not close to now", time.UnixMilli(ts))
	}

	want, err := GenSign("my-secret", ts)
	if err != nil {
		t.Fatalf("GenSign() error = %v", err)
	}
	if got := q.Get("sign"); got != want {
		t.Errorf("sign = %q, want %q", got, want)
	}

	// The sign must be URL-encoded in the raw query, since base64 can contain
	// '+', '/' and '='.
	if raw := u.RawQuery; !strings.Contains(raw, url.QueryEscape(want)) {
		t.Errorf("raw query %q does not contain encoded sign %q", raw, url.QueryEscape(want))
	}
}
