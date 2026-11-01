package models

import (
	"strings"
	"testing"
	"time"

	"github.com/bougou/alertmanager-webhook-adapter/pkg/models/templates"
)

func nodeDiskIOSaturationMessage() *AlertmanagerWebhookMessage {
	return &AlertmanagerWebhookMessage{
		Status:      "firing",
		Signature:   "test",
		GroupLabels: KV{"alertname": "NodeDiskIOSaturation"},
		MessageAt:   time.Date(2026, 11, 4, 16, 5, 21, 0, time.Local),
		Alerts: Alerts{
			{
				Status: "firing",
				Labels: KV{
					"alertname":          "NodeDiskIOSaturation",
					"severity":           "warning",
					"instance":           "10.0.0.10:9100",
					"device":             "sda",
					"mountpoint":         "/var",
					"fstype":             "ext4",
					"namespace":          "node-ns",
					"pod":                "node-exporter-pod",
					"container":          "node-exporter-container",
					"job":                "node-exporter-job",
					"endpoint":           "10.9.9.9:9100",
					"prometheus":         "monitoring/kps",
					"prometheus_replica": "prometheus-kps-0",
					"service":            "node-exporter-svc",
				},
				Annotations: KV{
					"description": "Disk IO queue (aqu-sq) is high on sda at 10.0.0.10:9100.\nThis symptom might indicate disk saturation.\n",
				},
				StartsAt: time.Date(2026, 10, 27, 18, 11, 20, 0, time.Local),
			},
		},
	}
}

func TestNodeDiskIOSaturationExtraLabelsLayout(t *testing.T) {
	msg := nodeDiskIOSaturationMessage()

	cases := []struct {
		name string
		src  string
		gap  string
	}{
		{"dingtalk", templates.DefaultTmplDingTalk, "disk saturation.\n>\n> **device**:"},
		{"dingtalk-zh", templates.DefaultTmplDingTalkZH, "disk saturation.\n>\n> **device**:"},
		{"feishu", templates.DefaultTmplFeishu, "disk saturation.\n**device**:"},
		{"feishu-zh", templates.DefaultTmplFeishuZH, "disk saturation.\n**device**:"},
		{"weixin", templates.DefaultTmplWeixin, "disk saturation.\n> <font color=\"comment\">device</font>:"},
		{"weixin-zh", templates.DefaultTmplWeixinZH, "disk saturation.\n> <font color=\"comment\">device</font>:"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := renderFeishuTmpl(t, c.src, msg, "prom.markdown")

			if !strings.Contains(out, c.gap) {
				t.Errorf("description and first extra label should be separated by a single line break, want %q in:\n%s", c.gap, out)
			}
			if strings.Contains(out, "disk saturation.\n\n") {
				t.Errorf("blank line found between description and extra labels:\n%s", out)
			}

			for _, want := range []string{"sda", "/var", "ext4"} {
				if !strings.Contains(out, want) {
					t.Errorf("expected extra label value %q in:\n%s", want, out)
				}
			}
			for _, unwanted := range []string{
				"node-ns",
				"node-exporter-pod",
				"node-exporter-container",
				"node-exporter-job",
				"10.9.9.9:9100",
				"monitoring/kps",
				"prometheus-kps-0",
				"node-exporter-svc",
			} {
				if strings.Contains(out, unwanted) {
					t.Errorf("did not expect hidden label value %q in:\n%s", unwanted, out)
				}
			}
		})
	}
}

func TestMixedGroupExtraLabelsAreFilteredPerAlert(t *testing.T) {
	msg := nodeDiskIOSaturationMessage()
	msg.Alerts = append(msg.Alerts, Alert{
		Status: "firing",
		Labels: KV{
			"alertname": "KubePodCrashLooping",
			"severity":  "warning",
			"instance":  "10.0.0.12:8080",
			"namespace": "workload-ns",
			"pod":       "api-0",
			"container": "api",
			"reason":    "CrashLoopBackOff",
		},
	})

	out := renderFeishuTmpl(t, templates.DefaultTmplFeishuZH, msg, "prom.markdown")

	if strings.Contains(out, "node-ns") {
		t.Errorf("node alert namespace should be hidden:\n%s", out)
	}
	if !strings.Contains(out, "workload-ns") {
		t.Errorf("workload alert namespace should be shown:\n%s", out)
	}
}

func TestResolvedAlertRendersExtraLabels(t *testing.T) {
	startsAt := time.Date(2026, 10, 27, 18, 11, 20, 0, time.Local)
	msg := &AlertmanagerWebhookMessage{
		Status:      "resolved",
		Signature:   "test",
		GroupLabels: KV{"alertname": "NodeDiskIOSaturation"},
		Alerts: Alerts{
			{
				Status:   "resolved",
				Labels:   KV{"alertname": "NodeDiskIOSaturation", "severity": "warning", "instance": "10.0.0.10:9100", "device": "sda"},
				StartsAt: startsAt,
				EndsAt:   startsAt.Add(time.Hour),
			},
		},
	}

	out := renderFeishuTmpl(t, templates.DefaultTmplFeishuZH, msg, "prom.markdown")
	if !strings.Contains(out, "**device**:") {
		t.Errorf("resolved alert should still render extra labels:\n%s", out)
	}
}

func TestExtraLabelValueTrailingNewlineDoesNotBreakList(t *testing.T) {
	msg := errorLogMessage()
	msg.Alerts[0].Labels["log"] = "first line\nsecond line\n"

	out := renderFeishuTmpl(t, templates.DefaultTmplFeishuZH, msg, "prom.markdown")

	if !strings.Contains(out, "second line</text_tag>\n**namespace**:") {
		t.Errorf("label value trailing newline should not add a blank line:\n%s", out)
	}
}

func kubeCPUOvercommitMessage(cluster string) *AlertmanagerWebhookMessage {
	labels := KV{"alertname": "KubeCPUOvercommit", "severity": "warning"}
	if cluster != "" {
		labels["cluster"] = cluster
	}
	return &AlertmanagerWebhookMessage{
		Status:      "firing",
		Signature:   "test",
		GroupLabels: KV{"alertname": "KubeCPUOvercommit"},
		Alerts: Alerts{
			{
				Status: "firing",
				Labels: labels,
				Annotations: KV{
					"description": "Cluster has overcommitted CPU resource requests for Pods by 1.99 CPU shares and cannot tolerate node failure.",
				},
				StartsAt: time.Date(2026, 10, 27, 17, 54, 4, 0, time.Local),
			},
		},
	}
}

func TestAlertInstanceFallsBackToAlertName(t *testing.T) {
	msg := kubeCPUOvercommitMessage("")

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
			if got := renderFeishuTmpl(t, c.src, &msg.Alerts[0], "__alertinstance"); got != "KubeCPUOvercommit" {
				t.Fatalf("expected alert name fallback, got %q", got)
			}
		})
	}

	out := renderFeishuTmpl(t, templates.DefaultTmplFeishuZH, msg, "prom.markdown")
	start := strings.Index(out, "**摘要**")
	end := strings.Index(out, "**详情**")
	if start < 0 || end < 0 || end <= start {
		t.Fatalf("could not locate summary section in:\n%s", out)
	}
	if summary := out[start:end]; !strings.Contains(summary, "KubeCPUOvercommit") {
		t.Errorf("summary should fall back to the alert name:\n%s", summary)
	}
}

func TestAlertInstancePrefersClusterOverAlertName(t *testing.T) {
	msg := kubeCPUOvercommitMessage("prod-cluster")

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
			if got := renderFeishuTmpl(t, c.src, &msg.Alerts[0], "__alertinstance"); got != "prod-cluster" {
				t.Fatalf("expected cluster fallback, got %q", got)
			}
		})
	}
}

func TestAlertDescriptionTrimsTrailingWhitespace(t *testing.T) {
	alert := &Alert{Annotations: KV{"description": "line1\nline2\n\n"}}
	alert.Annotations["description_en"] = "line1\nline2\n\n"
	alert.Annotations["description_cn"] = "line1\nline2\n\n"

	cases := []struct {
		name string
		src  string
	}{
		{"dingtalk", templates.DefaultTmplDingTalk},
		{"dingtalk-zh", templates.DefaultTmplDingTalkZH},
		{"feishu", templates.DefaultTmplFeishu},
		{"feishu-zh", templates.DefaultTmplFeishuZH},
		{"weixin", templates.DefaultTmplWeixin},
		{"weixin-zh", templates.DefaultTmplWeixinZH},
		{"weixinapp", templates.DefaultTmplWeixinapp},
		{"weixinapp-zh", templates.DefaultTmplWeixinappZH},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := renderFeishuTmpl(t, c.src, alert, "__alert_description")
			if got != "line1\nline2" {
				t.Fatalf("__alert_description = %q, want %q", got, "line1\nline2")
			}
		})
	}
}
