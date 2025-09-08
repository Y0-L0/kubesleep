package k8s

import (
	"log/slog"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s K8Simpl) GetSuspendableNamespace(namespace string) (kubesleep.SuspendableNamespace, error) {
	kubernetesNamespace, err := k8s.clientset.CoreV1().Namespaces().Get(k8s.ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return buildSuspendableNamespace(*kubernetesNamespace), err
}

func (k8s K8Simpl) GetSuspendableNamespaces() ([]kubesleep.SuspendableNamespace, error) {
	var result []kubesleep.SuspendableNamespace
	namespaces, err := k8s.clientset.CoreV1().Namespaces().List(k8s.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, ns := range namespaces.Items {
		result = append(result, buildSuspendableNamespace(ns))
	}
	return result, nil
}

func buildSuspendableNamespace(kubernetesNamespace corev1.Namespace) kubesleep.SuspendableNamespace {
	protected := false
	if kubernetesNamespace.ObjectMeta.Annotations != nil {
		_, protected = kubernetesNamespace.ObjectMeta.Annotations["kubesleep.xyz/do-not-suspend"]
	} else {
		slog.Debug("Namespace has no relevant annotations", "namespace", kubernetesNamespace.Name)
	}

	slog.Debug("namespace manifest", "protected", protected, "kubernetesNamespace", kubernetesNamespace)
	namespaceObj := kubesleep.NewSuspendableNamespace(
		kubernetesNamespace.Name,
		protected,
	)
	slog.Debug("parsed namespace", "namespace", namespaceObj)
	return namespaceObj

}
