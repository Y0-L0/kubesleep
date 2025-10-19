package k8s

import (
	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type K8Simpl struct {
	clientset *kubernetes.Clientset
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
	k8s.clientset, err = kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return k8s, nil
}
