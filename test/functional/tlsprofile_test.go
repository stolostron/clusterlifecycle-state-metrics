// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

//go:build functional
// +build functional

package functional

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	libgoconfig "github.com/stolostron/library-go/pkg/config"
)

const (
	apiServerName     = "cluster"
	metricsNamespace  = "multicluster-engine"
	metricsAppLabel   = "app=clusterlifecycle-state-metrics"
	metricsContainer  = "clusterlifecycle-state-metrics"
	podRestartTimeout = 120 * time.Second

	metricsHTTPSPort   = "https://metrics.localhost/metrics"
	telemetryHTTPSPort = "https://telemetry.localhost/metrics"
)

var _ = Describe("TLS Profile Functional Tests", func() {
	const (
		tlsHandshakeTimeout = 30 * time.Second
	)

	var (
		configClient configclient.Interface
	)

	BeforeEach(func() {
		var err error
		config, err := libgoconfig.LoadConfig("", kubeConfig, "")
		Expect(err).To(BeNil())

		configClient, err = configclient.NewForConfig(config)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		apiServer, err := configClient.ConfigV1().APIServers().Get(context.TODO(), apiServerName, metav1.GetOptions{})
		if err != nil {
			Fail(fmt.Sprintf("Failed to get APIServer CR: %v", err))
		}
		if apiServer.Spec.TLSSecurityProfile != nil {
			previous, err := currentMetricsContainer()
			Expect(err).To(BeNil())
			// Restore original TLS profile
			err = updateAPIServerProfile(configClient, nil)
			Expect(err).To(BeNil())
			waitForMetricsPodRestart(previous)
		}
	})

	Context("TLS Profile Configuration", func() {
		It("Should apply Intermediate TLS profile (TLS 1.2) to HTTPS endpoints", func() {
			// Create or update APIServer with Intermediate profile
			intermediateProfile := &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileIntermediateType,
			}

			previous, err := currentMetricsContainer()
			Expect(err).To(BeNil())

			err = updateAPIServerProfile(configClient, intermediateProfile)
			Expect(err).To(BeNil())

			klog.Info("Set TLS profile to Intermediate (TLS 1.2), waiting for pod restart...")
			waitForMetricsPodRestart(previous)

			// Verify both HTTPS endpoints honor the TLS 1.2 minimum
			By("Checking metrics endpoint on port 8443")
			Eventually(func() error {
				return verifyTLSEndpoint(metricsHTTPSPort, tls.VersionTLS12)
			}, podRestartTimeout, 5*time.Second).Should(Succeed())

			By("Checking telemetry endpoint on port 8444")
			Eventually(func() error {
				return verifyTLSEndpoint(telemetryHTTPSPort, tls.VersionTLS12)
			}, tlsHandshakeTimeout, 2*time.Second).Should(Succeed())

			By("Verifying TLS 1.2 clients are accepted")
			Eventually(func() error {
				return verifyTLS12Response(metricsHTTPSPort, true)
			}, tlsHandshakeTimeout, 2*time.Second).Should(Succeed())
			Eventually(func() error {
				return verifyTLS12Response(telemetryHTTPSPort, true)
			}, tlsHandshakeTimeout, 2*time.Second).Should(Succeed())
		})

		It("Should apply Modern TLS profile (TLS 1.3) to HTTPS endpoints", func() {
			// Create or update APIServer with Modern profile
			modernProfile := &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileModernType,
			}

			previous, err := currentMetricsContainer()
			Expect(err).To(BeNil())

			err = updateAPIServerProfile(configClient, modernProfile)
			Expect(err).To(BeNil())

			klog.Info("Set TLS profile to Modern (TLS 1.3), waiting for pod restart...")
			waitForMetricsPodRestart(previous)

			// Verify both HTTPS endpoints require TLS 1.3
			By("Checking metrics endpoint requires TLS 1.3")
			Eventually(func() error {
				return verifyTLSEndpoint(metricsHTTPSPort, tls.VersionTLS13)
			}, podRestartTimeout, 5*time.Second).Should(Succeed())

			By("Checking telemetry endpoint requires TLS 1.3")
			Eventually(func() error {
				return verifyTLSEndpoint(telemetryHTTPSPort, tls.VersionTLS13)
			}, tlsHandshakeTimeout, 2*time.Second).Should(Succeed())

			By("Verifying TLS 1.2 clients are rejected")
			Eventually(func() error {
				return verifyTLS12Response(metricsHTTPSPort, false)
			}, tlsHandshakeTimeout, 2*time.Second).Should(Succeed())
			Eventually(func() error {
				return verifyTLS12Response(telemetryHTTPSPort, false)
			}, tlsHandshakeTimeout, 2*time.Second).Should(Succeed())
		})

		It("Should apply Custom TLS profile with specific settings", func() {
			// Create or update APIServer with Custom profile
			customProfile := &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileCustomType,
				Custom: &configv1.CustomTLSProfile{
					TLSProfileSpec: configv1.TLSProfileSpec{
						MinTLSVersion: configv1.VersionTLS12,
						Ciphers: []string{
							"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
							"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
						},
					},
				},
			}

			previous, err := currentMetricsContainer()
			Expect(err).To(BeNil())

			err = updateAPIServerProfile(configClient, customProfile)
			Expect(err).To(BeNil())

			klog.Info("Set TLS profile to Custom (TLS 1.2 with specific ciphers), waiting for pod restart...")
			waitForMetricsPodRestart(previous)

			// Verify endpoints use TLS 1.2 minimum
			By("Checking custom TLS settings are applied")
			Eventually(func() error {
				return verifyTLSEndpoint(metricsHTTPSPort, tls.VersionTLS12)
			}, podRestartTimeout, 5*time.Second).Should(Succeed())

			// Verify the cipher suite is from our allowed list
			By("Checking cipher suite is from allowed list")
			Eventually(func() error {
				return verifyTLSCipher(metricsHTTPSPort, []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				})
			}, tlsHandshakeTimeout, 2*time.Second).Should(Succeed())

			By("Verifying TLS 1.2 clients are accepted")
			Eventually(func() error {
				return verifyTLS12Response(metricsHTTPSPort, true)
			}, tlsHandshakeTimeout, 2*time.Second).Should(Succeed())
		})
	})

	Context("TLS Profile Changes", func() {
		It("Should detect TLS profile changes and restart gracefully", func() {
			// Set initial profile to Intermediate
			intermediateProfile := &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileIntermediateType,
			}

			previous, err := currentMetricsContainer()
			Expect(err).To(BeNil())

			err = updateAPIServerProfile(configClient, intermediateProfile)
			Expect(err).To(BeNil())
			waitForMetricsPodRestart(previous)

			klog.Info("Set initial profile to Intermediate, waiting for stability...")

			// Verify Intermediate is active
			Eventually(func() error {
				return verifyTLSEndpoint(metricsHTTPSPort, tls.VersionTLS12)
			}, podRestartTimeout, 5*time.Second).Should(Succeed())

			By("Verifying TLS 1.2 clients are accepted")
			Eventually(func() error {
				return verifyTLS12Response(metricsHTTPSPort, true)
			}, tlsHandshakeTimeout, 2*time.Second).Should(Succeed())

			// Change to Modern profile
			modernProfile := &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileModernType,
			}

			previous, err = currentMetricsContainer()
			Expect(err).To(BeNil())

			klog.Info("Changing TLS profile to Modern...")
			err = updateAPIServerProfile(configClient, modernProfile)
			Expect(err).To(BeNil())
			waitForMetricsPodRestart(previous)

			// Verify Modern profile is now active (TLS 1.3)
			By("Verifying Modern profile is active after change")
			Eventually(func() error {
				return verifyTLSEndpoint(metricsHTTPSPort, tls.VersionTLS13)
			}, podRestartTimeout, 5*time.Second).Should(Succeed())

			By("Verifying TLS 1.2 clients are rejected")
			Eventually(func() error {
				return verifyTLS12Response(metricsHTTPSPort, false)
			}, tlsHandshakeTimeout, 2*time.Second).Should(Succeed())
		})
	})
})

// updateAPIServerProfile updates the APIServer CR with the specified TLS profile
func updateAPIServerProfile(configClient configclient.Interface, profile *configv1.TLSSecurityProfile) error {
	apiServer, err := configClient.ConfigV1().APIServers().Get(context.TODO(), apiServerName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	apiServer.Spec.TLSSecurityProfile = profile
	_, err = configClient.ConfigV1().APIServers().Update(context.TODO(), apiServer, metav1.UpdateOptions{})
	return err
}

// verifyTLSEndpoint connects to an endpoint and verifies the negotiated TLS
// version is at least minVersion. The client offers at most TLS 1.3.
func verifyTLSEndpoint(endpoint string, minVersion uint16) error {
	const maxVersion = tls.VersionTLS13

	// Create TLS config that records what was negotiated
	var negotiatedVersion uint16
	var negotiatedCipher uint16

	// #nosec G402 -- InsecureSkipVerify is acceptable in tests with self-signed certificates
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         minVersion,
		MaxVersion:         maxVersion,
		VerifyConnection: func(cs tls.ConnectionState) error {
			negotiatedVersion = cs.Version
			negotiatedCipher = cs.CipherSuite
			return nil
		},
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, endpoint)
	}

	// Verify TLS version
	if negotiatedVersion < minVersion {
		return fmt.Errorf("negotiated TLS version 0x%04x is below minimum 0x%04x",
			negotiatedVersion, minVersion)
	}

	if negotiatedVersion > maxVersion {
		return fmt.Errorf("negotiated TLS version 0x%04x is above maximum 0x%04x",
			negotiatedVersion, maxVersion)
	}

	klog.Infof("✓ %s: TLS version=0x%04x, cipher=0x%04x", endpoint, negotiatedVersion, negotiatedCipher)
	return nil
}

func tlsHandshakeRejected(err error) bool {
	var header tls.RecordHeaderError
	if errors.As(err, &header) {
		return true
	}

	var op *net.OpError
	if errors.As(err, &op) && op.Op == "remote error" {
		return true
	}

	// The client can also fail locally, before an alert arrives:
	// "tls: server selected unsupported protocol version 303".
	message := err.Error()
	return strings.Contains(message, "tls: protocol version") ||
		strings.Contains(message, "tls: handshake failure") ||
		strings.Contains(message, "tls: server selected unsupported")
}

// verifyTLS12Response verifies that a TLS 1.2 client to the endpoint is
// accepted or rejected, and that the HTTP response agrees with the error.
func verifyTLS12Response(endpoint string, accept bool) error {
	const tlsVersion = tls.VersionTLS12
	// #nosec G402 -- InsecureSkipVerify and low TLS versions are acceptable in tests
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tlsVersion,
		MaxVersion:         tlsVersion,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 10 * time.Second,
	}

	response, err := client.Get(endpoint)
	if response != nil {
		defer response.Body.Close()
	}

	// http.Client returns a response with a nil error, or an error with a nil
	// response. A TLS handshake failure is the latter. Both or neither means
	// the result cannot be attributed to the handshake.
	if err == nil && response == nil {
		return fmt.Errorf("TLS version 0x%04x returned neither a response nor an error", tlsVersion)
	}
	if err != nil && response != nil {
		return fmt.Errorf("TLS version 0x%04x returned both status %d and error: %v", tlsVersion, response.StatusCode, err)
	}

	if accept {
		if err != nil {
			return fmt.Errorf("expected TLS handshake to succeed with version 0x%04x, but it failed: %v", tlsVersion, err)
		}
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("expected TLS handshake to succeed with version 0x%04x, but it returned status %d", tlsVersion, response.StatusCode)
		}
		klog.Infof("✓ TLS version 0x%04x correctly accepted", tlsVersion)
		return nil
	}

	if err == nil {
		return fmt.Errorf("expected TLS handshake to fail with version 0x%04x, but it succeeded", tlsVersion)
	}
	if !tlsHandshakeRejected(err) {
		return fmt.Errorf("TLS version 0x%04x failed before the handshake: %w", tlsVersion, err)
	}
	klog.Infof("✓ TLS version 0x%04x correctly rejected: %v", tlsVersion, err)
	return nil
}

// verifyTLSCipher verifies that the server accepts every allowed cipher and
// rejects a cipher outside that list.
func verifyTLSCipher(endpoint string, allowedCiphers []uint16) error {
	for _, cipher := range allowedCiphers {
		// #nosec G402 -- InsecureSkipVerify is acceptable in tests with self-signed certificates
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			CipherSuites:       []uint16{cipher},
			MinVersion:         tls.VersionTLS12, // TLS 1.2 for cipher suite testing
			MaxVersion:         tls.VersionTLS12,
		}

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
			Timeout: 5 * time.Second,
		}

		resp, err := client.Get(endpoint)
		if err != nil {
			return fmt.Errorf("allowed cipher 0x%04x was rejected: %w", cipher, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK || resp.TLS == nil {
			return fmt.Errorf("allowed cipher 0x%04x returned status %d", cipher, resp.StatusCode)
		}
		klog.Infof("✓ Cipher 0x%04x is supported", cipher)
	}

	// Use a weak cipher that the custom and modern profiles should not offer.
	nonAllowedCipher := tls.TLS_RSA_WITH_AES_128_CBC_SHA
	if slices.Contains(allowedCiphers, nonAllowedCipher) {
		return fmt.Errorf("weak cipher 0x%04x chosen for testing but found in allow list", nonAllowedCipher)
	}

	// #nosec G402 -- InsecureSkipVerify is acceptable in tests with self-signed certificates
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		CipherSuites:       []uint16{nonAllowedCipher},
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS12,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(endpoint)
	if err == nil {
		resp.Body.Close()
		return fmt.Errorf("server accepted non-allowed cipher 0x%04x (status %d)", nonAllowedCipher, resp.StatusCode)
	}
	if !tlsHandshakeRejected(err) {
		return fmt.Errorf("non-allowed cipher 0x%04x failed before the handshake: %w", nonAllowedCipher, err)
	}
	klog.Infof("✓ Server correctly rejected non-allowed cipher 0x%04x", nonAllowedCipher)
	return nil
}

type metricsPodState struct {
	name        string
	uid         string
	containerID string
	ready       bool
	running     bool
	detail      string
}

// waitForMetricsPodRestart waits until the metrics container has exited and a
// replacement is ready. The process exits on a TLS profile change so it can
// reload TLS config. Kubelet backs off those restarts (10s, 20s, 40s, ... up
// to 300s) and this suite changes the profile often enough to exceed
// podRestartTimeout. Once the same pod's container has exited, delete it so
// the deployment starts a new pod immediately.
func waitForMetricsPodRestart(previous metricsPodState) {
	replacedPod := false
	Eventually(func(g Gomega) {
		current, err := currentMetricsContainer()
		if err == nil && !replacedPod && current.uid == previous.uid && !current.running {
			klog.Infof("Metrics container exited (%s); deleting pod to avoid kubelet restart backoff", current.detail)
			err := kubeClient.CoreV1().Pods(metricsNamespace).Delete(context.TODO(), current.name, metav1.DeleteOptions{})
			g.Expect(err).ToNot(HaveOccurred())
			replacedPod = true
		}
		g.Expect(err).To(BeNil())
		g.Expect(current.ready).To(BeTrue(), "metrics pod should be ready: %s", current.detail)
		g.Expect(current.containerID).NotTo(Equal(previous.containerID), "metrics container was not replaced: %s", current.detail)
	}, podRestartTimeout, 5*time.Second).Should(Succeed())
}

func currentMetricsContainer() (metricsPodState, error) {
	var state metricsPodState
	podList, err := kubeClient.CoreV1().Pods(metricsNamespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: metricsAppLabel,
	})
	if err != nil {
		return state, err
	}

	selected := -1
	for i := range podList.Items {
		if podList.Items[i].DeletionTimestamp != nil {
			continue
		}
		if selected >= 0 {
			return state, fmt.Errorf("found multiple pods in namespace %s with label %s", metricsNamespace, metricsAppLabel)
		}
		selected = i
	}
	if selected < 0 {
		return state, fmt.Errorf("no pod found in namespace %s with label %s", metricsNamespace, metricsAppLabel)
	}

	pod := &podList.Items[selected]
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name != metricsContainer {
			continue
		}
		waiting := ""
		if status.State.Waiting != nil {
			waiting = status.State.Waiting.Reason
		}
		terminated := ""
		if status.State.Terminated != nil {
			terminated = fmt.Sprintf("%s/%d", status.State.Terminated.Reason, status.State.Terminated.ExitCode)
		}
		return metricsPodState{
			name:        pod.Name,
			uid:         string(pod.UID),
			containerID: status.ContainerID,
			ready:       status.Ready && status.State.Running != nil,
			running:     status.State.Running != nil,
			detail: fmt.Sprintf("pod=%s containerID=%s ready=%t restarts=%d waiting=%s terminated=%s",
				pod.Name, status.ContainerID, status.Ready, status.RestartCount, waiting, terminated),
		}, nil
	}

	return state, fmt.Errorf("container %s not found in pod %s/%s", metricsContainer, pod.Namespace, pod.Name)
}
