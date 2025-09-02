package k8s

import (
	"log/slog"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s K8Simpl) GetStatefulSets(namespace string) (map[string]kubesleep.Suspendable, error) {
	statefulSets, err := k8s.clientset.AppsV1().
		StatefulSets(namespace).
		List(k8s.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	suspendables := map[string]kubesleep.Suspendable{}

	for _, statefulSet := range statefulSets.Items {
		suspend := k8s.suspendStatefulSet(namespace, statefulSet.Name)

		s := kubesleep.NewSuspendable(
			kubesleep.StatefulSet,
			statefulSet.Name,
			*statefulSet.Spec.Replicas,
			suspend,
		)
		slog.Debug("parsed Suspendable", "Suspendable", s)
		suspendables[s.Identifier()] = s
	}

	return suspendables, nil
}

func (k8s K8Simpl) suspendStatefulSet(namespace string, name string) func() error {
	return func() error {
		scalable, err := k8s.clientset.AppsV1().
			StatefulSets(namespace).
			GetScale(k8s.ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		scalable.Spec.Replicas = int32(0)

		_, err = k8s.clientset.AppsV1().StatefulSets(namespace).UpdateScale(
			k8s.ctx,
			scalable.Name,
			scalable,
			metav1.UpdateOptions{},
		)
		if err != nil {
			return err
		}
		slog.Info("Suspended Deployment", "name", scalable.Name)
		return nil
	}
}

func (k8s K8Simpl) ScaleStatefulSet(namespace string, name string, Replicas int32) error {
	scalable, err := k8s.clientset.AppsV1().
		StatefulSets(namespace).
		GetScale(k8s.ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	scalable.Spec.Replicas = Replicas
	_, err = k8s.clientset.AppsV1().StatefulSets(scalable.Namespace).UpdateScale(
		k8s.ctx,
		scalable.Name,
		scalable,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}

	slog.Info("Woke up StatefulSet", "name", name)
	return nil
}
