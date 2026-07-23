package k8s

import (
	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	kruiseclientset "github.com/openkruise/kruise-api/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type MixedClientset struct {
	*kubernetes.Clientset
	kruise *kruiseclientset.Clientset
}

type K8Simpl struct {
	clientset *MixedClientset
}

func NewK8S() (kubesleep.K8S, error) {
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	clientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	clientConfig.QPS = 10
	clientConfig.Burst = 100

	k8s := &K8Simpl{}
	k8sClientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	kruiseClientset, err := kruiseclientset.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	k8s.clientset = &MixedClientset{
		Clientset: k8sClientset,
		kruise:    kruiseClientset,
	}
	if err != nil {
		return nil, err
	}

	return k8s, nil
}
