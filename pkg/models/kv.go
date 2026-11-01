package models

import (
	"sort"
	"strings"
)

// Pair is a key/value string pair.
type Pair struct {
	Name, Value string
}

// Pairs is a list of key/value string pairs.
type Pairs []Pair

// Names returns a list of names of the pairs.
func (ps Pairs) Names() []string {
	ns := make([]string, 0, len(ps))
	for _, p := range ps {
		ns = append(ns, p.Name)
	}
	return ns
}

// Values returns a list of values of the pairs.
func (ps Pairs) Values() []string {
	vs := make([]string, 0, len(ps))
	for _, p := range ps {
		vs = append(vs, p.Value)
	}
	return vs
}

// KV is a set of key/value string pairs.
type KV map[string]string

// SortedPairs returns a sorted list of key/value pairs.
func (kv KV) SortedPairs() Pairs {
	var (
		pairs     = make([]Pair, 0, len(kv))
		keys      = make([]string, 0, len(kv))
		sortStart = 0
	)
	for k := range kv {
		if k == "alertname" {
			keys = append([]string{k}, keys...)
			sortStart = 1
		} else {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys[sortStart:])

	for _, k := range keys {
		pairs = append(pairs, Pair{k, kv[k]})
	}
	return pairs
}

// Remove returns a copy of the key/value set without the given keys.
func (kv KV) Remove(keys []string) KV {
	keySet := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keySet[k] = struct{}{}
	}

	res := KV{}
	for k, v := range kv {
		if _, ok := keySet[k]; !ok {
			res[k] = v
		}
	}
	return res
}

// Names returns the names of the label names in the LabelSet.
func (kv KV) Names() []string {
	return kv.SortedPairs().Names()
}

// Values returns a list of the values in the LabelSet.
func (kv KV) Values() []string {
	return kv.SortedPairs().Values()
}

// hiddenExtraLabels are omitted from the free-form extra label section. They
// are either already rendered as dedicated notification fields, duplicate
// annotation text, or Prometheus scrape/operator plumbing.
var hiddenExtraLabels = []string{
	"alertname", "alertname_cn", "alertname_custom",
	"severity",
	"alertinstance", "instance", "node", "nodename", "host", "hostname", "ip",
	"region", "zone", "product", "component",
	"description", "summary",
	"job", "endpoint", "prometheus", "prometheus_replica", "service",
}

// nodeScopedExtraLabels are meaningful for workload alerts but are scrape
// target metadata for node-scoped alerts.
var nodeScopedExtraLabels = []string{"namespace", "pod", "container"}

// Extra returns the labels worth rendering as free-form extra fields. Labels
// already shown as dedicated fields, annotation duplicates, and Prometheus
// scrape plumbing are omitted. Node-scoped alerts also omit workload identity
// labels.
func (kv KV) Extra() Pairs {
	hidden := make([]string, 0, len(hiddenExtraLabels)+len(nodeScopedExtraLabels)+1)
	hidden = append(hidden, hiddenExtraLabels...)
	if kv.isNodeScoped() {
		hidden = append(hidden, nodeScopedExtraLabels...)
	}
	// __alertinstance falls back to app, then cluster, then the alert name,
	// so the chosen fallback would otherwise be rendered twice.
	if !kv.hasInstanceLabel() {
		if kv["app"] != "" {
			hidden = append(hidden, "app")
		} else if kv["cluster"] != "" {
			hidden = append(hidden, "cluster")
		}
	}
	return kv.Remove(hidden).SortedPairs()
}

// isNodeScoped reports whether the alert describes a node rather than a
// workload. A node label alone is not enough: kube-state-metrics adds it to
// many workload alerts.
func (kv KV) isNodeScoped() bool {
	return strings.HasPrefix(kv["alertname"], "Node") ||
		kv["device"] != "" ||
		kv["mountpoint"] != ""
}

// hasInstanceLabel reports whether __alertinstance would pick one of its
// preferred labels before falling back to app.
func (kv KV) hasInstanceLabel() bool {
	for _, name := range []string{"alertinstance", "instance", "node", "nodename", "host", "hostname", "ip"} {
		if kv[name] != "" {
			return true
		}
	}
	return false
}
