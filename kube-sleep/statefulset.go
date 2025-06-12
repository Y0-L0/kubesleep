package kubesleep

import (
	"log/slog"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s k8simpl) getStatefulSets(namespace string) (map[string]suspendable, error) {
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

		s := suspendable{
			manifestType: "StatefulSet",
			name:         statefulSet.ObjectMeta.Name,
			replicas:     *statefulSet.Spec.Replicas,
			suspend:      suspend,
		}
		slog.Debug("parsed suspendable", "suspendable", s)
		suspendables[s.Identifier()] = s
	}

	return suspendables, nil
}

func (k8s k8simpl) suspendStatefulSet(statefulSet *appsv1.StatefulSet) error {
	replicas := int32(0)
	statefulSet.Spec.Replicas = &replicas
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

func (k8s k8simpl) scaleStatefulSet(namespace string, name string, replicas int32) error {
	statefulSet, err := k8s.clientset.AppsV1().StatefulSets(namespace).Get(
		k8s.ctx,
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		return err
	}
	statefulSet.Spec.Replicas = &replicas

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
