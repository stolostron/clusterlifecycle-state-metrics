// Copyright (c) 2020 Red Hat, Inc.

package options

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/pflag"

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

	Flags *pflag.FlagSet
}

func NewOptions() *Options {
	return &Options{
		Collectors:      koptions.CollectorSet{},
		MetricWhitelist: koptions.MetricSet{},
		MetricBlacklist: koptions.MetricSet{},
	}
}

func (o *Options) AddFlags() {
	o.Flags = pflag.NewFlagSet("", pflag.ExitOnError)
	// add klog flags
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
	o.Flags.AddGoFlagSet(klogFlags)
	if err := o.Flags.Lookup("logtostderr").Value.Set("true"); err != nil {
		panic(err)
	}
	o.Flags.Lookup("logtostderr").DefValue = "true"
	o.Flags.Lookup("logtostderr").NoOptDefVal = "true"

	o.Flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		o.Flags.PrintDefaults()
	}

	o.Flags.StringVar(&o.Apiserver, "apiserver", "", `The URL of the apiserver to use as a master`)
	o.Flags.StringVar(&o.Kubeconfig, "kubeconfig", "", "Absolute path to the kubeconfig file")
	o.Flags.BoolVarP(&o.Help, "help", "h", false, "Print Help text")
	o.Flags.IntVar(&o.HTTPPort, "http-port", 8080, `http Port to expose metrics on.`)
	o.Flags.IntVar(&o.HTTPSPort, "https-port", 8443, `https Port to expose metrics on.`)
	o.Flags.StringVar(&o.Host, "host", "0.0.0.0", `Host to expose metrics on.`)
	o.Flags.IntVar(&o.HTTPTelemetryPort, "http-telemetry-port", 8081, `http Port to expose openshift-state-metrics self metrics on.`)
	o.Flags.IntVar(&o.HTTPSTelemetryPort, "https-telemetry-port", 8444, `https Port to expose openshift-state-metrics self metrics on.`)
	o.Flags.StringVar(&o.TelemetryHost, "telemetry-host", "0.0.0.0", `Host to expose openshift-state-metrics self metrics on.`)
	o.Flags.StringVar(&o.TLSCrtFile, "tls-crt-file", "", `TLS certificate file path.`)
	o.Flags.StringVar(&o.TLSKeyFile, "tls-key-file", "", `TLS key file path.`)
	o.Flags.Var(&o.Collectors, "collectors", fmt.Sprintf("Comma-separated list of collectors to be enabled. Defaults to %q", &DefaultCollectors))
	o.Flags.Var(&o.Namespaces, "namespace", fmt.Sprintf("Comma-separated list of namespaces to be enabled. Defaults to %q", &DefaultNamespaces))
	o.Flags.Var(&o.MetricWhitelist, "metric-whitelist", "Comma-separated list of metrics to be exposed. The whitelist and blacklist are mutually exclusive.")
	o.Flags.Var(&o.MetricBlacklist, "metric-blacklist", "Comma-separated list of metrics not to be enabled. The whitelist and blacklist are mutually exclusive.")
	o.Flags.BoolVarP(&o.Version, "version", "", false, "openshift-state-metrics build version information")

	o.Flags.BoolVar(&o.EnableGZIPEncoding, "enable-gzip-encoding", false, "Gzip responses when requested by clients via 'Accept-Encoding: gzip' header.")
}

func (o *Options) Parse() error {
	klog.Infof("OS.args: %v", os.Args)
	err := o.Flags.Parse(os.Args)
	return err
}

func (o *Options) Usage() {
	o.Flags.Usage()
}
