package kubesleep

import (
	"context"
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type k8simpl struct {
	clientset *kubernetes.Clientset
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewK8simpl() (*k8simpl, error) {
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	clientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	k8s := &k8simpl{}
	k8s.clientset, err = kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	k8s.ctx, k8s.cancel = context.WithCancel(k8s.ctx)

	return k8s, nil
}

func (k8s k8simpl) suspendableNamespace(namespace string) (suspendableNamespace, error) {
	kubernetesNamespace, err := k8s.clientset.CoreV1().Namespaces().Get(k8s.ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	namespaceObj := suspendableNamespaceImpl{
		name:         namespace,
		_suspendable: true,
	}
	if kubernetesNamespace.ObjectMeta.Annotations != nil {
		_, found := kubernetesNamespace.ObjectMeta.Annotations["kubesleep.xyz/do-not-suspend"]
		namespaceObj._suspendable = !found
	} else {
		slog.Debug("Namespace has no annotations")
	}

	slog.Debug("namespace manifest", "kubernetesNamespace", kubernetesNamespace)
	slog.Info("parsed namespace", "namespace", namespaceObj)
	return &namespaceObj, err
}
