// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package options

import (
	"flag"
	"fmt"
	"os"

	"k8s.io/klog/v2"
	koptions "k8s.io/kube-state-metrics/pkg/options"
)

type Options struct {
	Apiserver          string
	Kubeconfig         string
	Help               bool
	HTTPPort           int
	HTTPSPort          int
	Host               string
	HTTPTelemetryPort  int
	HTTPSTelemetryPort int
	TelemetryHost      string
	TLSCrtFile         string
	TLSKeyFile         string
	Collectors         koptions.CollectorSet
	Namespaces         koptions.NamespaceList
	MetricBlacklist    koptions.MetricSet
	MetricWhitelist    koptions.MetricSet
	Version            bool

	EnableGZIPEncoding bool
}

func NewOptions() *Options {
	return &Options{
		Collectors:      koptions.CollectorSet{},
		MetricWhitelist: koptions.MetricSet{},
		MetricBlacklist: koptions.MetricSet{},
	}
}

func (o *Options) AddFlags() {
	klog.Info("Start add args")
	klog.InitFlags(flag.CommandLine)
	if err := flag.Lookup("logtostderr").Value.Set("true"); err != nil {
		panic(err)
	}
	flag.Lookup("logtostderr").DefValue = "true"

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&o.Apiserver, "apiserver", "", `The URL of the apiserver to use as a server`)
	flag.StringVar(&o.Kubeconfig, "csm-kubeconfig", "", "Absolute path to the kubeconfig file")
	flag.BoolVar(&o.Help, "help", false, "Print Help text")
	flag.IntVar(&o.HTTPPort, "http-port", 8080, `http Port to expose metrics on.`)
	flag.IntVar(&o.HTTPSPort, "https-port", 8443, `https Port to expose metrics on.`)
	flag.StringVar(&o.Host, "host", "0.0.0.0", `Host to expose metrics on.`)
	flag.IntVar(&o.HTTPTelemetryPort, "http-telemetry-port", 8081, `http Port to expose openshift-state-metrics self metrics on.`)
	flag.IntVar(&o.HTTPSTelemetryPort, "https-telemetry-port", 8444, `https Port to expose openshift-state-metrics self metrics on.`)
	flag.StringVar(&o.TelemetryHost, "telemetry-host", "0.0.0.0", `Host to expose openshift-state-metrics self metrics on.`)
	flag.StringVar(&o.TLSCrtFile, "tls-crt-file", "", `TLS certificate file path.`)
	flag.StringVar(&o.TLSKeyFile, "tls-key-file", "", `TLS key file path.`)
	flag.Var(&o.Collectors, "collectors", fmt.Sprintf("Comma-separated list of collectors to be enabled. Defaults to %q", &DefaultCollectors))
	flag.Var(&o.Namespaces, "namespace", fmt.Sprintf("Comma-separated list of namespaces to be enabled. Defaults to %q", &DefaultNamespaces))
	flag.Var(&o.MetricWhitelist, "metric-whitelist", "Comma-separated list of metrics to be exposed. The whitelist and blacklist are mutually exclusive.")
	flag.Var(&o.MetricBlacklist, "metric-blacklist", "Comma-separated list of metrics not to be enabled. The whitelist and blacklist are mutually exclusive.")
	flag.BoolVar(&o.Version, "version", false, "openshift-state-metrics build version information")

	flag.BoolVar(&o.EnableGZIPEncoding, "enable-gzip-encoding", false, "Gzip responses when requested by clients via 'Accept-Encoding: gzip' header.")
	klog.Info("End add args")
}

func (o *Options) Parse() {
	if flag.Parsed() {
		return
	}
	flag.Parse()
}

func (o *Options) Usage() {
	flag.Usage()
}
