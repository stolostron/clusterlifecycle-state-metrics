// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	/* #nosec */
	kubeConfigFileBasic = "../../../test/unit/tmp/envtest/kubeconfig/basic"
	kubeConfigFileT     = "../../../test/unit/tmp/envtest/kubeconfig/token"
	kubeConfigFileCerts = "../../../test/unit/tmp/envtest/kubeconfig/certs"
)

var envTests map[string]*envtest.Environment

func TestMain(m *testing.M) {
	exitCode := m.Run()
	tearDownEnvTests()
	os.Exit(exitCode)
}

func setupEnvTest(t *testing.T) (envTest *envtest.Environment) {
	return setupEnvTestByName(t.Name())
}

func setupEnvTestByName(name string) (envTest *envtest.Environment) {
	if envTests == nil {
		envTests = make(map[string]*envtest.Environment)
	}
	if e, ok := envTests[name]; ok {
		return e
	}
	//Create an envTest
	envTest = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join(
				"..",
				"..",
				"test",
				"functional",
				"resources",
				"crds",
				"managedclusters_crd.yaml",
			),
			filepath.Join(
				"..",
				"..",
				"test",
				"functional",
				"resources",
				"crds",
				"managedclusteraddons_crd.yaml",
			),
			filepath.Join(
				"..",
				"..",
				"test",
				"functional",
				"resources",
				"crds",
				"manifestworks_crd.yaml",
			),
		},
	}
	envTests[name] = envTest
	_, err := envTest.Start()
	if err != nil {
		panic(err)
	}
	return envTest
}

func tearDownEnvTests() {
	for name := range envTests {
		tearDownByName(name)
	}
}

func tearDownByName(name string) {
	if e, ok := envTests[name]; ok {
		e.Stop()
	}
}
