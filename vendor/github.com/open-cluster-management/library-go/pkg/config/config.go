package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

//LoadConfig loads the kubeconfig and returns a *rest.Config
//url: The url of the server
//kubeconfig: The path of the kubeconfig, if empty the KUBECONFIG environment variable will be used.
//context: The context to use to search the *rest.Config in the kubeconfig file,
// if empty the current-context will be used.
//Search in the following order:
// provided kubeconfig path, KUBECONFIG environment variable, in the cluster, in the user home directory.
//If the context is not provided and the url is not provided, it returns a *rest.Config the for the current-context.
//If the context is not provided but the url provided, it returns a *rest.Config for the server identified by the url.
//If the context is provided, it returns a *rest.Config for the provided context.
func LoadConfig(
	url,
	kubeconfig,
	context string,
) (*rest.Config, error) {
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}
	klog.V(5).Infof("Kubeconfig path %s\n", kubeconfig)
	// If we have an explicit indication of where the kubernetes config lives, read that.
	if kubeconfig != "" {
		return configFromFile(url, kubeconfig, context)
	}
	// If not, try the in-cluster config.
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	// If no in-cluster config, try the default location in the user's home directory.
	if usr, err := user.Current(); err == nil {
		klog.V(5).Infof("clientcmd.BuildConfigFromFlags for url %s using %s\n",
			url,
			filepath.Join(usr.HomeDir,
				".kube",
				"config"))
		return configFromFile(url, filepath.Join(usr.HomeDir, ".kube", "config"), context)
	}

	return nil, fmt.Errorf("could not create a valid kubeconfig")

}

func configFromFile(url, kubeconfig, context string) (*rest.Config, error) {
	if context == "" {
		// klog.V(5).Infof("clientcmd.BuildConfigFromFlags with %s and %s", url, kubeconfig)
		// Retreive the config for the current context
		if url == "" {
			config, err := clientcmd.LoadFromFile(kubeconfig)
			if err != nil {
				return nil, err
			}
			return clientcmd.NewDefaultClientConfig(
				*config,
				&clientcmd.ConfigOverrides{}).ClientConfig()
		}
		return clientcmd.BuildConfigFromFlags(url, kubeconfig)
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}
