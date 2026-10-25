package api

import (
	"net/url"
	"strings"
	"testing"
)

func TestRedactURL(t *testing.T) {
	u, err := url.Parse("http://example.com/webhook/send/?channel_type=feishu&msg_type=markdown&token=s3cr3t&secret=shh")
	if err != nil {
		t.Fatalf("parse URL: %v", err)
	}

	got := redactURL(u)
	for _, credential := range []string{"s3cr3t", "shh"} {
		if strings.Contains(got, credential) {
			t.Errorf("redactURL() = %q, still contains %q", got, credential)
		}
	}

	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse redacted URL: %v", err)
	}
	q := parsed.Query()
	if got := q.Get("token"); got != redactedValue {
		t.Errorf("token = %q, want %q", got, redactedValue)
	}
	if got := q.Get("secret"); got != redactedValue {
		t.Errorf("secret = %q, want %q", got, redactedValue)
	}
	if got := q.Get("channel_type"); got != "feishu" {
		t.Errorf("channel_type = %q, want %q", got, "feishu")
	}
	if got := q.Get("msg_type"); got != "markdown" {
		t.Errorf("msg_type = %q, want %q", got, "markdown")
	}
}
