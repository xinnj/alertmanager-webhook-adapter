package feishu

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func newTestBot(t *testing.T, token, secret string) (*FeishuGroupBot, *map[string]interface{}) {
	t.Helper()
	return newTestBotWithResponse(t, token, secret, http.StatusOK, `{"code":0,"msg":"success"}`)
}

func newTestBotWithResponse(t *testing.T, token, secret string, status int, response string) (*FeishuGroupBot, *map[string]interface{}) {
	t.Helper()

	body := make(map[string]interface{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.WriteHeader(status)
		if response != "" {
			_, _ = w.Write([]byte(response))
		}
	}))
	t.Cleanup(srv.Close)

	bot := NewFeishuGroupBotWithSecret(token, secret)
	bot.addr = srv.URL
	return bot, &body
}

func TestSendReturnsErrorOnBusinessError(t *testing.T) {
	bot, _ := newTestBotWithResponse(t, "my-token", "my-secret", http.StatusOK,
		`{"code":19021,"msg":"sign match fail or timestamp is not within one hour from current time"}`)
	msg := &Msg{MsgType: MsgTypeText, Content: &Content{Text: "hello"}}

	err := bot.Send(msg)
	if err == nil {
		t.Fatal("Send() error = nil, want error")
	}
	for _, want := range []string{"19021", "sign match fail"} {
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
	bot, _ := newTestBotWithResponse(t, "my-token", "my-secret", http.StatusOK, `<html>bad gateway</html>`)
	msg := &Msg{MsgType: MsgTypeText, Content: &Content{Text: "hello"}}

	err := bot.Send(msg)
	if err == nil {
		t.Fatal("Send() error = nil, want error")
	}
	if strings.Contains(err.Error(), "bad gateway") {
		t.Errorf("error %q unexpectedly includes the response body", err)
	}
}

func TestSendWithSecretAddsTimestampAndSign(t *testing.T) {
	bot, body := newTestBot(t, "my-token", "my-secret")
	msg := &Msg{MsgType: MsgTypeText, Content: &Content{Text: "hello"}}

	if err := bot.Send(msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	ts, ok := (*body)["timestamp"].(string)
	if !ok {
		t.Fatalf("request body has no string timestamp: %v", *body)
	}
	timestamp, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		t.Fatalf("timestamp %q is not an integer: %v", ts, err)
	}

	sign, ok := (*body)["sign"].(string)
	if !ok {
		t.Fatalf("request body has no sign: %v", *body)
	}

	want, err := GenSign("my-secret", timestamp)
	if err != nil {
		t.Fatalf("GenSign() error = %v", err)
	}
	if sign != want {
		t.Errorf("sign = %q, want %q", sign, want)
	}
}

func TestSendWithoutSecretLeavesBodyUnsigned(t *testing.T) {
	bot, body := newTestBot(t, "my-token", "")
	msg := &Msg{MsgType: MsgTypeText, Content: &Content{Text: "hello"}}

	if err := bot.Send(msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if _, ok := (*body)["timestamp"]; ok {
		t.Errorf("unsigned body unexpectedly has timestamp: %v", *body)
	}
	if _, ok := (*body)["sign"]; ok {
		t.Errorf("unsigned body unexpectedly has sign: %v", *body)
	}
}

func TestSendErrorDoesNotLeakCredentials(t *testing.T) {
	bot := NewFeishuGroupBotWithSecret("my-token", "my-secret")
	bot.addr = "http://127.0.0.1:1"
	msg := &Msg{MsgType: MsgTypeText, Content: &Content{Text: "hello"}}

	err := bot.Send(msg)
	if err == nil {
		t.Fatal("Send() error = nil, want error")
	}

	for _, credential := range []string{"my-token", "my-secret"} {
		if strings.Contains(err.Error(), credential) {
			t.Errorf("error %q contains %q", err, credential)
		}
	}
}

func TestGenSignFixedVector(t *testing.T) {
	got, err := GenSign("my-secret", 1700000000)
	if err != nil {
		t.Fatalf("GenSign() error = %v", err)
	}
	const want = "I3DorsRQAITp7wWwxm5u7O9Ca8T+zLwkNSN7C16z0yQ="
	if got != want {
		t.Errorf("GenSign() = %q, want %q", got, want)
	}
}
