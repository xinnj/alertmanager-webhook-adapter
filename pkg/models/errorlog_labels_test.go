package models

import (
	"strings"
	"testing"
	"time"

	"github.com/bougou/alertmanager-webhook-adapter/pkg/models/templates"
)

func errorLogMessage() *AlertmanagerWebhookMessage {
	return &AlertmanagerWebhookMessage{
		Status:      "firing",
		Signature:   "test",
		Receiver:    "log",
		MessageAt:   time.Date(2026, 10, 29, 16, 17, 11, 0, time.Local),
		GroupLabels: KV{"alertname": "ErrorLog"},
		Alerts: Alerts{
			{
				Status: "firing",
				Labels: KV{
					"alertname": "ErrorLog",
					"severity":  "warning",
					"namespace": "next",
					"app":       "bvpronext-ai-agent",
					"time":      "2026-10-20T16:07:07.098+08:00",
					"log":       "531947af-75ac-421a-895c-958c09071473|/usr/src/app/src/middlewares/error.js:45:17||SyntaxError: JSON5: invalid character ',' at 2:1",
				},
				Annotations: KV{"description": "Log level with ERROR in log"},
				StartsAt:    time.Date(2026, 10, 20, 16, 7, 7, 0, time.Local),
			},
		},
	}
}

// ErrorLog alerts have no instance-like label. The alert instance must fall
// back to the app label so the summary and detail are not left empty.
func TestAlertInstanceFallsBackToApp(t *testing.T) {
	msg := errorLogMessage()

	cases := []struct {
		name string
		src  string
	}{
		{"default", templates.DefaultTmpl},
		{"default-zh", templates.DefaultTmplZH},
		{"dingtalk", templates.DefaultTmplDingTalk},
		{"dingtalk-zh", templates.DefaultTmplDingTalkZH},
		{"feishu", templates.DefaultTmplFeishu},
		{"feishu-zh", templates.DefaultTmplFeishuZH},
		{"weixin", templates.DefaultTmplWeixin},
		{"weixin-zh", templates.DefaultTmplWeixinZH},
		{"weixinapp", templates.DefaultTmplWeixinapp},
		{"weixinapp-zh", templates.DefaultTmplWeixinappZH},
		{"slack", templates.DefaultTmplSlack},
		{"slack-zh", templates.DefaultTmplSlackZH},
		{"discord-webhook", templates.DefaultTmplDiscordWebhook},
		{"discord-webhook-zh", templates.DefaultTmplDiscordWebhookZH},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := renderFeishuTmpl(t, c.src, &msg.Alerts[0], "__alertinstance"); got != "bvpronext-ai-agent" {
				t.Errorf("expected app fallback as alert instance, got %q", got)
			}
		})
	}
}

// ErrorLog alerts carry their payload in the log label, not in the
// instance-like labels the templates normally render. The notification must
// still include the extra labels so operators can see the error.
func TestErrorLogNotificationIncludesExtraLabels(t *testing.T) {
	msg := errorLogMessage()

	channels := []struct {
		name string
		src  string
	}{
		{"dingtalk", templates.DefaultTmplDingTalk},
		{"dingtalk-zh", templates.DefaultTmplDingTalkZH},
		{"feishu", templates.DefaultTmplFeishu},
		{"feishu-zh", templates.DefaultTmplFeishuZH},
		{"weixin", templates.DefaultTmplWeixin},
		{"weixin-zh", templates.DefaultTmplWeixinZH},
	}

	for _, c := range channels {
		t.Run(c.name, func(t *testing.T) {
			out := renderFeishuTmpl(t, c.src, msg, "prom.markdown")

			for _, want := range []string{
				"namespace",
				"bvpronext-ai-agent",
				"2026-10-20T16:07:07.098+08:00",
				"SyntaxError: JSON5",
			} {
				if !strings.Contains(out, want) {
					t.Errorf("expected %q in %s notification, got:\n%s", want, c.name, out)
				}
			}

			// Labels already rendered as dedicated fields must not be repeated.
			if strings.Contains(out, "**alertname**") {
				t.Errorf("did not expect alertname label in %s notification, got:\n%s", c.name, out)
			}
		})
	}
}
