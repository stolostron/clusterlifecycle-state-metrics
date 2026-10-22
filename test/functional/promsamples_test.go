// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package functional

import (
	"reflect"
	"testing"
)

func TestVectorSamplesIgnoreOrderAndName(t *testing.T) {
	rule := `{"status":"success","data":{"resultType":"vector","result":[` +
		`{"metric":{"__name__":"acm_managed_cluster_worker_cores:max","hub_cluster_id":"hub","managed_cluster_id":"local_cluster_id"},"value":[10,"0"]},` +
		`{"metric":{"__name__":"acm_managed_cluster_worker_cores:max","hub_cluster_id":"hub","managed_cluster_id":"hive_cluster_id"},"value":[10,"0"]},` +
		`{"metric":{"__name__":"acm_managed_cluster_worker_cores:max","hub_cluster_id":"hub","managed_cluster_id":"import_cluster_id"},"value":[10,"0"]},` +
		`{"metric":{"__name__":"acm_managed_cluster_worker_cores:max","hub_cluster_id":"hub","managed_cluster_id":"empty-cluster"},"value":[10,"0"]}` +
		`]}}`
	expr := `{"status":"success","data":{"resultType":"vector","result":[` +
		`{"metric":{"hub_cluster_id":"hub","managed_cluster_id":"hive_cluster_id"},"value":[11,"0"]},` +
		`{"metric":{"hub_cluster_id":"hub","managed_cluster_id":"import_cluster_id"},"value":[11,"0"]},` +
		`{"metric":{"hub_cluster_id":"hub","managed_cluster_id":"empty-cluster"},"value":[11,"0"]},` +
		`{"metric":{"hub_cluster_id":"hub","managed_cluster_id":"local_cluster_id"},"value":[11,"0"]}` +
		`]}}`

	ruleSamples, err := vectorSamples(rule)
	if err != nil {
		t.Fatal(err)
	}
	exprSamples, err := vectorSamples(expr)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ruleSamples, exprSamples) {
		t.Fatalf("series differ by order or __name__:\n rule %#v\n expr %#v", ruleSamples, exprSamples)
	}

	exposition := "" +
		"# HELP acm_managed_cluster_worker_cores cores\n" +
		"# TYPE acm_managed_cluster_worker_cores gauge\n" +
		"acm_managed_cluster_worker_cores{managed_cluster_id=\"empty-cluster\",hub_cluster_id=\"hub\"} 0\n" +
		"acm_managed_cluster_worker_cores{hub_cluster_id=\"hub\",managed_cluster_id=\"local_cluster_id\"} 0\n" +
		"acm_managed_cluster_worker_cores{hub_cluster_id=\"hub\",managed_cluster_id=\"import_cluster_id\"} 0\n" +
		"acm_managed_cluster_worker_cores{hub_cluster_id=\"hub\",managed_cluster_id=\"hive_cluster_id\"} 0\n"
	expected, err := expositionSamples(exposition)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ruleSamples, expected) {
		t.Fatalf("exposition samples:\n got  %#v\n want %#v", expected, ruleSamples)
	}
}

func TestVectorSamplesPreserveDifferentValues(t *testing.T) {
	stale := `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"hub_cluster_id":"hub","managed_cluster_id":"hive_cluster_id"},"value":[1,"0"]}]}}`
	current := `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"managed_cluster_id":"hive_cluster_id","hub_cluster_id":"hub"},"value":[2,"5"]}]}}`

	staleSamples, err := vectorSamples(stale)
	if err != nil {
		t.Fatal(err)
	}
	currentSamples, err := vectorSamples(current)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(staleSamples, currentSamples) {
		t.Fatalf("stale and current samples matched: %#v", staleSamples)
	}
}
