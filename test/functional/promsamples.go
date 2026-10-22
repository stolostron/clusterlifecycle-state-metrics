// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package functional

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type promVector struct {
	Data struct {
		Result []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// vectorSamples reads a Prometheus instant-query body into label-set to value.
// __name__ and the sample timestamp are omitted so a recording rule can be
// compared with the aggregation that produced it.
func vectorSamples(body string) (map[string]string, error) {
	var parsed promVector
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		return nil, err
	}
	samples := make(map[string]string, len(parsed.Data.Result))
	for _, series := range parsed.Data.Result {
		if len(series.Value) < 2 {
			return nil, fmt.Errorf("sample missing value for %v", series.Metric)
		}
		value, ok := series.Value[1].(string)
		if !ok {
			return nil, fmt.Errorf("sample value for %v is %T, want string", series.Metric, series.Value[1])
		}
		samples[seriesKey(series.Metric)] = value
	}
	return samples, nil
}

// expositionSamples reads Prometheus text exposition lines into the same
// label-set map as vectorSamples. HELP and TYPE comments are ignored.
func expositionSamples(exposition string) (map[string]string, error) {
	samples := map[string]string{}
	for line := range strings.SplitSeq(exposition, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		nameEnd := strings.IndexByte(line, '{')
		braceEnd := strings.LastIndexByte(line, '}')
		if nameEnd < 0 || braceEnd < nameEnd {
			return nil, fmt.Errorf("metric line missing labels: %s", line)
		}
		labels := map[string]string{}
		for pair := range strings.SplitSeq(line[nameEnd+1:braceEnd], ",") {
			eq := strings.IndexByte(pair, '=')
			if eq < 0 {
				return nil, fmt.Errorf("metric label missing value: %s", pair)
			}
			labels[pair[:eq]] = strings.Trim(pair[eq+1:], `"`)
		}
		samples[seriesKey(labels)] = strings.TrimSpace(line[braceEnd+1:])
	}
	return samples, nil
}

func seriesKey(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for key := range labels {
		if key == "__name__" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, key := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteString(labels[key])
	}
	return b.String()
}
