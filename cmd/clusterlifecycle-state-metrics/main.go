// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package main

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	koptions "k8s.io/kube-state-metrics/pkg/options"
	"k8s.io/kube-state-metrics/pkg/whiteblacklist"
	workv1 "open-cluster-management.io/api/work/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/collectors"
	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/controllers"
	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/options"
	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/version"
)

const (
	leaderConfigMapName = "clusterlifecycle-state-metrics-lock"
	metricsPath         = "/metrics"
	healthzPath         = "/healthz"

	hubTypeMCE              = "mce"
	hubTypeACM              = "acm"
	hubTypeStolostronEngine = "stolostron-engine"
	hubTypeStolostron       = "stolostron"

	configMapName = "clusterlifecycle-state-metrics-config"
)

var (
	opts   *options.Options
	scheme = runtime.NewScheme()
)

// promLogger implements promhttp.Logger
type promLogger struct{}

func init() {
	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling flag.Parse().
	zapopts := zap.Options{}
	zapopts.BindFlags(flag.CommandLine)

	logger := zap.New(zap.UseFlagOptions(&zapopts))
	logf.SetLogger(logger)

	opts = options.NewOptions()
	opts.AddFlags()

	// init schema
	_ = clientgoscheme.AddToScheme(scheme)
	_ = workv1.AddToScheme(scheme)
}

func (pl promLogger) Println(v ...interface{}) {
	klog.Error(v...)
}

func main() {
	opts.Parse()

	if opts.Version {
		fmt.Printf("%#v\n", version.GetVersion())
		os.Exit(0)
	}

	if opts.Help {
		opts.Usage()
		os.Exit(0)
	}
	start(opts)
}

func start(opts *options.Options) {
	ctx := context.TODO()
	config, err := clientcmd.BuildConfigFromFlags(opts.Apiserver, opts.Kubeconfig)
	if err != nil {
		klog.Fatalf("cannot create config: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("cannot create kubeClient: %v", err)
	}

	timestampMetricsEnabled, err := isTimestampMetricsEnabled(ctx, kubeClient)
	if err != nil {
		klog.Fatalf("cannot determine if timestamp metrics should be enabled: %v", err)
	}

	if timestampMetricsEnabled {
		go startControllers(opts)
	}

	collectorBuilder := collectors.NewBuilder(ctx)
	collectorBuilder.WithRestConfig(config).
		WithKubeclient(kubeClient).
		WithTimestampMetricsEnabled(timestampMetricsEnabled)
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

	switch opts.HubType {
	case hubTypeMCE, hubTypeACM, hubTypeStolostronEngine, hubTypeStolostron:
		collectorBuilder.WithHubType(opts.HubType)
	case "":
		// do nothing if not specified
	default:
		klog.Fatal(fmt.Errorf("invalid hub type %q", opts.HubType))
	}

	whiteBlackList, err := whiteblacklist.New(opts.MetricWhitelist, opts.MetricBlacklist)
	if err != nil {
		klog.Fatal(err)
	}
	if err := whiteBlackList.Parse(); err != nil {
		klog.Fatal(err)
	}

	klog.Infof("metric white-black listing: %v", whiteBlackList.Status())

	collectorBuilder.WithWhiteBlackList(whiteBlackList)

	ocmMetricsRegistry := prometheus.NewRegistry()
	if err := ocmMetricsRegistry.Register(collectors.ResourcesPerScrapeMetric); err != nil {
		panic(err)
	}
	if err := ocmMetricsRegistry.Register(collectors.ScrapeErrorTotalMetric); err != nil {
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

func startControllers(opts *options.Options) {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: opts.ControllerMetricsAddress,
		},
		// WebhookServer:          webhookServer,
		// HealthProbeBindAddress: probeAddr,
		LeaderElection:   opts.EnableLeaderElection,
		LeaderElectionID: "ecaf1259.my.domain",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		LeaderElectionReleaseOnCancel: true,
		Cache: cache.Options{
			DefaultTransform: func(obj interface{}) (interface{}, error) {
				mw, ok := obj.(*workv1.ManifestWork)
				if !ok {
					return obj, nil
				}

				// Transform the object to only include metadata and conditions
				return &workv1.ManifestWork{
					ObjectMeta: metav1.ObjectMeta{
						Name:              mw.Name,
						Namespace:         mw.Namespace,
						Annotations:       mw.Annotations,
						Labels:            mw.Labels,
						CreationTimestamp: mw.CreationTimestamp,
						UID:               mw.UID,
					},
					Status: workv1.ManifestWorkStatus{
						Conditions: mw.Status.Conditions,
					},
				}, nil
			},
		},
	})
	if err != nil {
		logf.Log.Error(err, "unable to start manager")
		os.Exit(1)
	} // speeds up voluntary leader transitions as the new leader don't have to wait

	now := time.Now()
	if err = controllers.NewManifestworkReconciler(mgr.GetClient(), now).SetupWithManager(mgr); err != nil {
		logf.Log.Error(err, "unable to create controller", "controller", "ManifestWork")
		os.Exit(1)
	}
	logf.Log.Info("The manifestwork controller start time has been set", "startTime", now)

	// +kubebuilder:scaffold:builder

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logf.Log.Error(err, "problem running manager")
		os.Exit(1)
	}
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
		s := &http.Server{
			Addr:      listenAddress,
			Handler:   mux,
			TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		}

		klog.Infof("Starting clusterlifecycle-state-metrics self metrics server: %s", listenAddress)
		klog.Infof("Listening https: %s", listenAddress)
		go func() { log.Fatal(s.ListenAndServeTLS(tlsCrtFile, tlsKeyFile)) }()
	}
	// Address to listen on for web interface and telemetry
	listenAddress := net.JoinHostPort(host, strconv.Itoa(httpPort))
	s := &http.Server{
		Addr:    listenAddress,
		Handler: mux,
	}

	klog.Infof("Starting clusterlifecycle-state-metrics self metrics server: %s", listenAddress)

	klog.Infof("Listening http: %s", listenAddress)

	log.Fatal(s.ListenAndServe())
}

func serveMetrics(collectors []collectors.MetricsCollector,
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
             <h1>ACM Metrics</h1>
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
		s := &http.Server{
			Addr:      listenAddress,
			Handler:   mux,
			TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		}

		klog.Infof("Starting metrics server: %s", listenAddress)
		klog.Infof("Listening https: %s", listenAddress)
		go func() { log.Fatal(s.ListenAndServeTLS(tlsCrtFile, tlsKeyFile)) }()
	}
	// Address to listen on for web interface and telemetry
	listenAddress := net.JoinHostPort(host, strconv.Itoa(httpPort))
	s := &http.Server{
		Addr:    listenAddress,
		Handler: mux,
	}

	klog.Infof("Starting metrics server: %s", listenAddress)

	klog.Infof("Listening http: %s", listenAddress)
	log.Fatal(s.ListenAndServe())
}

type metricHandler struct {
	collectors         []collectors.MetricsCollector
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
		c.WriteAll(w)
	}

	// In case we gziped the response, we have to close the writer.
	if closer, ok := writer.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			panic(err)
		}
	}
}

// Check the ConfigMap to see if the timestamp-metrics should be enabled
func isTimestampMetricsEnabled(ctx context.Context, clientset *kubernetes.Clientset) (bool, error) {
	namespace, err := GetComponentNamespace()
	if err != nil {
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("failed to get namespace: %v", err)
		}
		// the local test will run into here
		klog.Info("serviceaccount namespace file does not exist, enable timestamp metrics")
		return true, nil
	}

	// Get the ConfigMap
	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to get ConfigMap: %v", err)
	}

	// Check if collect-timestamp-metrics is "enable"
	value, exists := cm.Data["collect-timestamp-metrics"]
	if !exists {
		return false, nil
	}

	return value == "Enable", nil
}

func GetComponentNamespace() (string, error) {
	nsBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "multicluster-engine", err
	}
	return string(nsBytes), nil
}
