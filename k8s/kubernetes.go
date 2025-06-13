package k8s

import (
	"context"
	"log/slog"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type K8Simpl struct {
	clientset *kubernetes.Clientset
	ctx       context.Context
	cancel    context.CancelFunc
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

	k8s := &K8Simpl{}
	k8s.clientset, err = kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	k8s.ctx, k8s.cancel = context.WithCancel(k8s.ctx)

	return k8s, nil
}

func (k8s K8Simpl) GetSuspendableNamespace(namespace string) (kubesleep.SuspendableNamespace, error) {
	kubernetesNamespace, err := k8s.clientset.CoreV1().Namespaces().Get(k8s.ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	suspendable := true
	if kubernetesNamespace.ObjectMeta.Annotations != nil {
		_, found := kubernetesNamespace.ObjectMeta.Annotations["kubesleep.xyz/do-not-suspend"]
		suspendable = !found
	} else {
		slog.Debug("Namespace has no relevant annotations")
	}

	slog.Debug("namespace manifest", "suspendable", suspendable, "kubernetesNamespace", kubernetesNamespace)
	namespaceObj := kubesleep.NewSuspendableNamespace(
		namespace,
		suspendable,
	)
	slog.Info("parsed namespace", "namespace", namespaceObj)
	return namespaceObj, err
}
