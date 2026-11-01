package models

import (
	"reflect"
	"testing"
)

func TestKVExtra(t *testing.T) {
	cases := []struct {
		name   string
		labels KV
		want   []string
	}{
		{
			name: "error log hides app because it is the instance fallback",
			labels: KV{
				"alertname": "ErrorLog",
				"severity":  "warning",
				"namespace": "next",
				"app":       "bvpronext-ai-agent",
				"time":      "2026-10-20T16:07:07.098+08:00",
				"log":       "SyntaxError",
			},
			want: []string{"log", "namespace", "time"},
		},
		{
			name: "node disk hides workload and scrape labels",
			labels: KV{
				"alertname":          "NodeDiskIOSaturation",
				"severity":           "warning",
				"instance":           "10.0.0.10:9100",
				"device":             "sda",
				"mountpoint":         "/var",
				"fstype":             "ext4",
				"job":                "node-exporter",
				"endpoint":           "10.9.9.9:9100",
				"prometheus":         "monitoring/kps",
				"prometheus_replica": "prometheus-kps-0",
				"service":            "node-exporter-svc",
				"namespace":          "monitoring",
				"pod":                "node-exporter-pod",
				"container":          "node-exporter",
			},
			want: []string{"device", "fstype", "mountpoint"},
		},
		{
			name: "device implies node scope on non-Node alertname",
			labels: KV{
				"alertname": "CPUHigh",
				"severity":  "warning",
				"instance":  "10.0.0.11:9100",
				"device":    "sda",
				"namespace": "workload",
				"pod":       "api-0",
				"app":       "api",
			},
			want: []string{"app", "device"},
		},
		{
			name: "workload alert keeps namespace pod and container",
			labels: KV{
				"alertname":  "KubePodCrashLooping",
				"severity":   "warning",
				"instance":   "10.0.0.12:8080",
				"namespace":  "app",
				"pod":        "api-0",
				"container":  "api",
				"app":        "api",
				"reason":     "CrashLoopBackOff",
				"job":        "kube-state-metrics",
				"endpoint":   "10.0.0.12:8080",
				"prometheus": "monitoring/kps",
				"service":    "kube-state-metrics",
			},
			want: []string{"app", "container", "namespace", "pod", "reason"},
		},
		{
			name: "all labels hidden",
			labels: KV{
				"alertname": "TargetDown",
				"severity":  "warning",
				"instance":  "10.0.0.13:9100",
			},
			want: []string{},
		},
		{
			name: "cluster is hidden when it is the instance fallback",
			labels: KV{
				"alertname": "KubeCPUOvercommit",
				"severity":  "warning",
				"cluster":   "prod",
			},
			want: []string{},
		},
		{
			name: "cluster stays when app is the instance fallback",
			labels: KV{
				"alertname": "ErrorLog",
				"severity":  "warning",
				"app":       "api",
				"cluster":   "prod",
				"log":       "error",
			},
			want: []string{"cluster", "log"},
		},
		{
			name:   "nil labels",
			labels: nil,
			want:   []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.labels.Extra().Names()
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("Extra() names = %v, want %v", got, c.want)
			}
		})
	}
}

func TestKVExtraDoesNotMutateLabels(t *testing.T) {
	labels := KV{
		"alertname": "NodeDiskIOSaturation",
		"severity":  "warning",
		"device":    "sda",
		"namespace": "monitoring",
		"pod":       "node-exporter-pod",
	}
	before := make(KV, len(labels))
	for k, v := range labels {
		before[k] = v
	}

	_ = labels.Extra()

	if !reflect.DeepEqual(labels, before) {
		t.Fatalf("Extra() mutated labels: got %v, want %v", labels, before)
	}
}
