package k8s

import (
	"log/slog"

	kubesleep "github.com/Y0-L0/kubesleep/kube-sleep"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s K8Simpl) GetStatefulSets(namespace string) (map[string]kubesleep.Suspendable, error) {
	statefulSets, err := k8s.clientset.AppsV1().
		StatefulSets(namespace).
		List(k8s.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	suspendables := map[string]suspendable{}

	for _, statefulSet := range statefulSets.Items {
		suspend := func() error {
			return k8s.suspendStatefulSet(&statefulSet)
		}

		s := kubesleep.NewSuspendable(
			kubesleep.StatefulSet,
			statefulSet.ObjectMeta.Name,
			*statefulSet.Spec.Replicas,
			suspend,
		)
		slog.Debug("parsed Suspendable", "Suspendable", s)
		suspendables[s.Identifier()] = s
	}

	return suspendables, nil
}

func (k8s K8Simpl) suspendStatefulSet(statefulSet *appsv1.StatefulSet) error {
	Replicas := int32(0)
	statefulSet.Spec.Replicas = &Replicas
	_, err := k8s.clientset.AppsV1().StatefulSets(statefulSet.Namespace).Update(
		k8s.ctx,
		statefulSet,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}
	slog.Info("Suspended StatefulSet", "name", statefulSet.Name)
	return nil
}

func (k8s K8Simpl) ScaleStatefulSet(namespace string, name string, Replicas int32) error {
	statefulSet, err := k8s.clientset.AppsV1().StatefulSets(namespace).Get(
		k8s.ctx,
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		return err
	}
	statefulSet.Spec.Replicas = &Replicas

	_, err = k8s.clientset.AppsV1().StatefulSets(statefulSet.Namespace).Update(
		k8s.ctx,
		statefulSet,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}

	slog.Info("Woke up StatefulSet", "name", name)
	return nil
}
