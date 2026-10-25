package utils

import (
	"errors"
	"net/url"
	"strings"
	"testing"
)

func TestSanitizeRequestErrorRemovesURL(t *testing.T) {
	err := &url.Error{
		Op:  "Post",
		URL: "https://open.feishu.cn/open-apis/bot/v2/hook/secret-token?sign=abc",
		Err: errors.New("boom"),
	}

	got := SanitizeRequestError(err).Error()

	for _, secret := range []string{"secret-token", "sign=abc", "https://"} {
		if strings.Contains(got, secret) {
			t.Errorf("sanitized error %q contains %q", got, secret)
		}
	}
	for _, want := range []string{"Post", "boom"} {
		if !strings.Contains(got, want) {
			t.Errorf("sanitized error %q does not contain %q", got, want)
		}
	}
}

func TestSanitizeRequestErrorPassesThroughOtherErrors(t *testing.T) {
	err := errors.New("plain error")

	if got := SanitizeRequestError(err); got != err {
		t.Errorf("SanitizeRequestError() = %v, want original error", got)
	}
}
