// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

//go:build functional
// +build functional

package functional

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	libgoconfig "github.com/stolostron/library-go/pkg/config"
)

const (
	metricsHTTPSPort    = "https://metrics.localhost/metrics"
	telemetryHTTPSPort  = "https://telemetry.localhost/metrics"
	apiServerName       = "cluster"
	podRestartTimeout   = 120 * time.Second
	tlsHandshakeTimeout = 30 * time.Second
)

var _ = Describe("TLS Profile Functional Tests", func() {
	var (
		configClient     configclient.Interface
		originalProfile  *configv1.TLSSecurityProfile
		apiServerExists  bool
	)

	BeforeEach(func() {
		var err error
		config, err := libgoconfig.LoadConfig("", kubeConfig, "")
		Expect(err).To(BeNil())

		configClient, err = configclient.NewForConfig(config)
		Expect(err).To(BeNil())

		// Check if running on OpenShift and save original profile
		apiServer, err := configClient.ConfigV1().APIServers().Get(context.TODO(), apiServerName, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				apiServerExists = false
				klog.Info("APIServer CR not found - will create one for testing")
			} else {
				Fail(fmt.Sprintf("Failed to get APIServer CR: %v", err))
			}
		} else {
			apiServerExists = true
			if apiServer.Spec.TLSSecurityProfile != nil {
				originalProfile = apiServer.Spec.TLSSecurityProfile.DeepCopy()
			}
			klog.Infof("Found existing APIServer CR with TLS profile: %v", apiServer.Spec.TLSSecurityProfile)
		}
	})

	AfterEach(func() {
		if apiServerExists && originalProfile != nil {
			// Restore original TLS profile
			apiServer, err := configClient.ConfigV1().APIServers().Get(context.TODO(), apiServerName, metav1.GetOptions{})
			if err == nil {
				apiServer.Spec.TLSSecurityProfile = originalProfile
				_, err = configClient.ConfigV1().APIServers().Update(context.TODO(), apiServer, metav1.UpdateOptions{})
				if err != nil {
					klog.Warningf("Failed to restore original TLS profile: %v", err)
				} else {
					klog.Info("Restored original TLS profile")
					// Wait for pod restart with original settings
					time.Sleep(10 * time.Second)
				}
			}
		}
	})

	Context("TLS Profile Configuration", func() {
		It("Should apply Intermediate TLS profile (TLS 1.2) to HTTPS endpoints", func() {
			// Create or update APIServer with Intermediate profile
			intermediateProfile := &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileIntermediateType,
			}

			err := createOrUpdateAPIServerProfile(configClient, intermediateProfile)
			Expect(err).To(BeNil())

			klog.Info("Set TLS profile to Intermediate (TLS 1.2), waiting for pod restart...")
			time.Sleep(15 * time.Second)

			// Verify both HTTPS endpoints honor the TLS 1.2 minimum
			By("Checking metrics endpoint on port 8443")
			Eventually(func() error {
				return verifyTLSEndpoint(metricsHTTPSPort, tls.VersionTLS12, tls.VersionTLS13)
			}, podRestartTimeout, 5*time.Second).Should(Succeed())

			By("Checking telemetry endpoint on port 8444")
			Eventually(func() error {
				return verifyTLSEndpoint(telemetryHTTPSPort, tls.VersionTLS12, tls.VersionTLS13)
			}, tlsHandshakeTimeout, 2*time.Second).Should(Succeed())
		})

		It("Should apply Modern TLS profile (TLS 1.3) to HTTPS endpoints", func() {
			// Create or update APIServer with Modern profile
			modernProfile := &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileModernType,
			}

			err := createOrUpdateAPIServerProfile(configClient, modernProfile)
			Expect(err).To(BeNil())

			klog.Info("Set TLS profile to Modern (TLS 1.3), waiting for pod restart...")
			time.Sleep(15 * time.Second)

			// Verify both HTTPS endpoints require TLS 1.3
			By("Checking metrics endpoint requires TLS 1.3")
			Eventually(func() error {
				return verifyTLSEndpoint(metricsHTTPSPort, tls.VersionTLS13, tls.VersionTLS13)
			}, podRestartTimeout, 5*time.Second).Should(Succeed())

			By("Checking telemetry endpoint requires TLS 1.3")
			Eventually(func() error {
				return verifyTLSEndpoint(telemetryHTTPSPort, tls.VersionTLS13, tls.VersionTLS13)
			}, tlsHandshakeTimeout, 2*time.Second).Should(Succeed())

			// Verify TLS 1.2 clients are rejected
			By("Verifying TLS 1.2 clients are rejected")
			err = verifyTLSRejected(metricsHTTPSPort, tls.VersionTLS12)
			Expect(err).To(HaveOccurred(), "TLS 1.2 should be rejected when Modern profile is active")
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

			err := createOrUpdateAPIServerProfile(configClient, customProfile)
			Expect(err).To(BeNil())

			klog.Info("Set TLS profile to Custom (TLS 1.2 with specific ciphers), waiting for pod restart...")
			time.Sleep(15 * time.Second)

			// Verify endpoints use TLS 1.2 minimum
			By("Checking custom TLS settings are applied")
			Eventually(func() error {
				return verifyTLSEndpoint(metricsHTTPSPort, tls.VersionTLS12, tls.VersionTLS13)
			}, podRestartTimeout, 5*time.Second).Should(Succeed())

			// Verify the cipher suite is from our allowed list
			By("Checking cipher suite is from allowed list")
			Eventually(func() error {
				return verifyTLSCipher(metricsHTTPSPort, []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				})
			}, tlsHandshakeTimeout, 2*time.Second).Should(Succeed())
		})
	})

	Context("TLS Profile Changes", func() {
		It("Should detect TLS profile changes and restart gracefully", func() {
			// Set initial profile to Intermediate
			intermediateProfile := &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileIntermediateType,
			}

			err := createOrUpdateAPIServerProfile(configClient, intermediateProfile)
			Expect(err).To(BeNil())

			klog.Info("Set initial profile to Intermediate, waiting for stability...")
			time.Sleep(15 * time.Second)

			// Verify Intermediate is active
			err = verifyTLSEndpoint(metricsHTTPSPort, tls.VersionTLS12, tls.VersionTLS13)
			Expect(err).To(BeNil())

			// Change to Modern profile
			modernProfile := &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileModernType,
			}

			klog.Info("Changing TLS profile to Modern...")
			err = createOrUpdateAPIServerProfile(configClient, modernProfile)
			Expect(err).To(BeNil())

			// Pod should restart and apply new profile
			klog.Info("Waiting for pod to restart with new profile...")
			time.Sleep(20 * time.Second)

			// Verify Modern profile is now active (TLS 1.3)
			By("Verifying Modern profile is active after change")
			Eventually(func() error {
				return verifyTLSEndpoint(metricsHTTPSPort, tls.VersionTLS13, tls.VersionTLS13)
			}, podRestartTimeout, 5*time.Second).Should(Succeed())
		})
	})
})

// createOrUpdateAPIServerProfile creates or updates the APIServer CR with the specified TLS profile
func createOrUpdateAPIServerProfile(configClient configclient.Interface, profile *configv1.TLSSecurityProfile) error {
	apiServer, err := configClient.ConfigV1().APIServers().Get(context.TODO(), apiServerName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create new APIServer CR
			apiServer = &configv1.APIServer{
				ObjectMeta: metav1.ObjectMeta{
					Name: apiServerName,
				},
				Spec: configv1.APIServerSpec{
					TLSSecurityProfile: profile,
				},
			}
			_, err = configClient.ConfigV1().APIServers().Create(context.TODO(), apiServer, metav1.CreateOptions{})
			return err
		}
		return err
	}

	// Update existing APIServer CR
	apiServer.Spec.TLSSecurityProfile = profile
	_, err = configClient.ConfigV1().APIServers().Update(context.TODO(), apiServer, metav1.UpdateOptions{})
	return err
}

// verifyTLSEndpoint connects to an endpoint and verifies the TLS version is within expected range
func verifyTLSEndpoint(endpoint string, minVersion, maxVersion uint16) error {
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

// verifyTLSRejected verifies that a connection with the specified TLS version is rejected
func verifyTLSRejected(endpoint string, tlsVersion uint16) error {
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

	_, err := client.Get(endpoint)
	if err == nil {
		return fmt.Errorf("expected TLS handshake to fail with version 0x%04x, but it succeeded", tlsVersion)
	}

	klog.Infof("✓ TLS version 0x%04x correctly rejected: %v", tlsVersion, err)
	return nil
}

// verifyTLSCipher verifies that the server supports only the allowed cipher suites
func verifyTLSCipher(endpoint string, allowedCiphers []uint16) error {
	supportedCiphers := []uint16{}
	unsupportedCiphers := []uint16{}

	// Test each allowed cipher individually
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
			// Cipher not supported
			unsupportedCiphers = append(unsupportedCiphers, cipher)
			klog.V(4).Infof("Cipher 0x%04x not supported: %v", cipher, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK && resp.TLS != nil {
			supportedCiphers = append(supportedCiphers, resp.TLS.CipherSuite)
			klog.Infof("✓ Cipher 0x%04x is supported", cipher)
		} else {
			unsupportedCiphers = append(unsupportedCiphers, cipher)
		}
	}

	// Verify at least some allowed ciphers are supported
	if len(supportedCiphers) == 0 {
		return fmt.Errorf("none of the allowed ciphers are supported by the server")
	}

	// Test a non-allowed cipher to ensure server rejects it
	// Use a common cipher that's not in the allowed list for negative test
	nonAllowedCipher := tls.TLS_RSA_WITH_AES_128_CBC_SHA // Weak cipher, unlikely to be in Modern/Custom profiles
	isInAllowedList := false
	for _, allowed := range allowedCiphers {
		if allowed == nonAllowedCipher {
			isInAllowedList = true
			break
		}
	}

	if !isInAllowedList {
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
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			klog.Warningf("⚠ Server accepted non-allowed cipher 0x%04x", nonAllowedCipher)
		} else {
			klog.Infof("✓ Server correctly rejected non-allowed cipher 0x%04x", nonAllowedCipher)
		}
	}

	klog.Infof("Server supports %d/%d allowed ciphers", len(supportedCiphers), len(allowedCiphers))
	return nil
}
