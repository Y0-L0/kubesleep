package k8s

import (
	"context"
	"log/slog"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s K8Simpl) getKruiseStatefulSets(ctx context.Context, namespace string) (map[string]kubesleep.Suspendable, error) {
	statefulSets, err := k8s.clientset.kruise.AppsV1beta1().
		StatefulSets(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	suspendables := map[string]kubesleep.Suspendable{}

	for _, statefulSet := range statefulSets.Items {
		var suspend func(context.Context) error
		if statefulSet.Spec.Replicas != nil && *statefulSet.Spec.Replicas == 0 {
			suspend = k8s.noopSuspendKruiseStatefulSet(namespace, statefulSet.Name)
		} else {
			suspend = k8s.suspendKruiseStatefulSet(namespace, statefulSet.Name)
		}

		s := kubesleep.NewSuspendable(
			kubesleep.KruiseStatefulSet,
			statefulSet.Name,
			*statefulSet.Spec.Replicas,
			suspend,
		)
		slog.Debug("parsed Suspendable", "Suspendable", s, "namespace", namespace)
		suspendables[s.Identifier()] = s
	}

	return suspendables, nil
}

func (k8s K8Simpl) noopSuspendKruiseStatefulSet(namespace, name string) func(context.Context) error {
	return func(ctx context.Context) error {
		slog.Debug("KruiseStatefulSet already at 0 replicas; skipping suspend", "namespace", namespace, "name", name)
		return nil
	}
}

func (k8s K8Simpl) suspendKruiseStatefulSet(namespace string, name string) func(context.Context) error {
	return func(ctx context.Context) error {
		scalable, err := k8s.clientset.kruise.AppsV1beta1().
			StatefulSets(namespace).
			GetScale(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		scalable.Spec.Replicas = int32(0)

		_, err = k8s.clientset.kruise.AppsV1beta1().StatefulSets(namespace).UpdateScale(
			ctx,
			scalable.Name,
			scalable,
			metav1.UpdateOptions{},
		)
		if err != nil {
			return err
		}
		slog.Info("Suspended KruiseStatefulSet", "name", scalable.Name, "namespace", namespace)
		return nil
	}
}

func (k8s K8Simpl) scaleKruiseStatefulSet(ctx context.Context, namespace string, name string, Replicas int32) error {
	scalable, err := k8s.clientset.kruise.AppsV1beta1().
		StatefulSets(namespace).
		GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	scalable.Spec.Replicas = Replicas
	_, err = k8s.clientset.kruise.AppsV1beta1().StatefulSets(scalable.Namespace).UpdateScale(
		ctx,
		scalable.Name,
		scalable,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}

	slog.Info("Woke up KruiseStatefulSet", "name", name, "namespace", namespace)
	return nil
}
