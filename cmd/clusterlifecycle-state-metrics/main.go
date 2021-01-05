// Copyright (c) 2020 Red Hat, Inc.

package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/open-cluster-management/clusterlifecycle-state-metrics/pkg/version"
	kcollectors "k8s.io/kube-state-metrics/pkg/collector"
	koptions "k8s.io/kube-state-metrics/pkg/options"
	"k8s.io/kube-state-metrics/pkg/whiteblacklist"

	ocollectors "github.com/open-cluster-management/clusterlifecycle-state-metrics/pkg/collectors"
	"github.com/open-cluster-management/clusterlifecycle-state-metrics/pkg/options"
)

const (
	metricsPath = "/metrics"
	healthzPath = "/healthz"
)

// promLogger implements promhttp.Logger
type promLogger struct{}

func (pl promLogger) Println(v ...interface{}) {
	klog.Error(v...)
}

func main() {
	opts := options.NewOptions()
	opts.AddFlags()

	err := opts.Parse()
	if err != nil {
		klog.Fatalf("Error: %s", err)
	}

	if opts.Version {
		fmt.Printf("%#v\n", version.GetVersion())
		os.Exit(0)
	}

	if opts.Help {
		opts.Usage()
		os.Exit(0)
	}

	collectorBuilder := ocollectors.NewBuilder(context.TODO())
	collectorBuilder.WithApiserver(opts.Apiserver).WithKubeConfig(opts.Kubeconfig)
	if len(opts.Collectors) == 0 {
		klog.Info("Using default collectors")
		collectorBuilder.WithEnabledCollectors(options.DefaultCollectors.AsSlice())
	} else {
		collectorBuilder.WithEnabledCollectors(opts.Collectors.AsSlice())
	}

	if len(opts.Namespaces) == 0 {
		klog.Info("Using all namespace")
		collectorBuilder.WithNamespaces(koptions.DefaultNamespaces)
	} else {
		if opts.Namespaces.IsAllNamespaces() {
			klog.Info("Using all namespace")
		} else {
			klog.Infof("Using %s namespaces", opts.Namespaces)
		}
		collectorBuilder.WithNamespaces(opts.Namespaces)
	}

	whiteBlackList, err := whiteblacklist.New(opts.MetricWhitelist, opts.MetricBlacklist)
	if err != nil {
		klog.Fatal(err)
	}

	klog.Infof("metric white- blacklisting: %v", whiteBlackList.Status())

	collectorBuilder.WithWhiteBlackList(whiteBlackList)

	ocmMetricsRegistry := prometheus.NewRegistry()
	if err := ocmMetricsRegistry.Register(ocollectors.ResourcesPerScrapeMetric); err != nil {
		panic(err)
	}
	if err := ocmMetricsRegistry.Register(ocollectors.ScrapeErrorTotalMetric); err != nil {
		panic(err)
	}
	if err := ocmMetricsRegistry.Register(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{})); err != nil {
		panic(err)
	}
	if err := ocmMetricsRegistry.Register(prometheus.NewGoCollector()); err != nil {
		panic(err)
	}
	go telemetryServer(ocmMetricsRegistry, opts.TelemetryHost, opts.HTTPTelemetryPort, opts.HTTPSTelemetryPort, opts.TLSCrtFile, opts.TLSKeyFile)

	collectors := collectorBuilder.Build()

	serveMetrics(collectors, opts.Host, opts.HTTPPort, opts.HTTPSPort, opts.TLSCrtFile, opts.TLSKeyFile, opts.EnableGZIPEncoding)
}

func telemetryServer(
	registry prometheus.Gatherer,
	host string,
	httpPort int,
	httpsPort int,
	tlsCrtFile string,
	tlsKeyFile string,
) {

	mux := http.NewServeMux()

	// Add metricsPath
	mux.Handle(metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{ErrorLog: promLogger{}}))
	// Add index
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`<html>
             <head><title>openshift-State-Metrics Metrics Server</title></head>
             <body>
             <h1>openshift-State-Metrics Metrics</h1>
			 <ul>
             <li><a href='` + metricsPath + `'>metrics</a></li>
			 </ul>
             </body>
             </html>`)); err != nil {
			panic(err)
		}
	})
	if tlsCrtFile != "" && tlsKeyFile != "" {
		// Address to listen on for web interface and telemetry
		listenAddress := net.JoinHostPort(host, strconv.Itoa(httpsPort))

		klog.Infof("Starting clusterlifecycle-state-metrics self metrics server: %s", listenAddress)
		klog.Infof("Listening https: %s", listenAddress)
		go func() { log.Fatal(http.ListenAndServeTLS(listenAddress, tlsCrtFile, tlsKeyFile, mux)) }()
	}
	// Address to listen on for web interface and telemetry
	listenAddress := net.JoinHostPort(host, strconv.Itoa(httpPort))

	klog.Infof("Starting clusterlifecycle-state-metrics self metrics server: %s", listenAddress)

	klog.Infof("Listening http: %s", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, mux))
}

func serveMetrics(collectors []*kcollectors.Collector,
	host string,
	httpPort int,
	httpsPort int,
	tlsCrtFile string,
	tlsKeyFile string,
	enableGZIPEncoding bool) {

	mux := http.NewServeMux()

	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	// Add metricsPath
	mux.Handle(metricsPath, &metricHandler{collectors, enableGZIPEncoding})
	// Add healthzPath
	mux.HandleFunc(healthzPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		if _, err := w.Write([]byte("ok")); err != nil {
			panic(err)
		}
	})
	// Add index
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`<html>
             <head><title>Open Cluster Managementt Metrics Server</title></head>
             <body>
             <h1>OCM Metrics</h1>
			 <ul>
             <li><a href='` + metricsPath + `'>metrics</a></li>
             <li><a href='` + healthzPath + `'>healthz</a></li>
			 </ul>
             </body>
             </html>`)); err != nil {
			panic(err)
		}
	})
	if tlsCrtFile != "" && tlsKeyFile != "" {
		// Address to listen on for web interface and telemetry
		listenAddress := net.JoinHostPort(host, strconv.Itoa(httpsPort))

		klog.Infof("Starting metrics server: %s", listenAddress)
		klog.Infof("Listening https: %s", listenAddress)
		go func() { log.Fatal(http.ListenAndServeTLS(listenAddress, tlsCrtFile, tlsKeyFile, mux)) }()
	}
	// Address to listen on for web interface and telemetry
	listenAddress := net.JoinHostPort(host, strconv.Itoa(httpPort))

	klog.Infof("Starting metrics server: %s", listenAddress)

	klog.Infof("Listening http: %s", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, mux))
}

type metricHandler struct {
	collectors         []*kcollectors.Collector
	enableGZIPEncoding bool
}

func (m *metricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resHeader := w.Header()
	var writer io.Writer = w

	resHeader.Set("Content-Type", `text/plain; version=`+"0.0.4")

	if m.enableGZIPEncoding {
		// Gzip response if requested. Taken from
		// github.com/prometheus/client_golang/prometheus/promhttp.decorateWriter.
		reqHeader := r.Header.Get("Accept-Encoding")
		parts := strings.Split(reqHeader, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "gzip" || strings.HasPrefix(part, "gzip;") {
				writer = gzip.NewWriter(writer)
				resHeader.Set("Content-Encoding", "gzip")
			}
		}
	}

	for _, c := range m.collectors {
		c.Collect(w)
	}

	// In case we gziped the response, we have to close the writer.
	if closer, ok := writer.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			panic(err)
		}
	}
}
