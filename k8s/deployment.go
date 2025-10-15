package k8s

import (
	"context"
	"log/slog"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s K8Simpl) getDeployments(ctx context.Context, namespace string) (map[string]kubesleep.Suspendable, error) {
	deployments, err := k8s.clientset.AppsV1().
		Deployments(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	suspendables := map[string]kubesleep.Suspendable{}

	for _, deployment := range deployments.Items {
		var suspend func() error
		if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == 0 {
			suspend = k8s.noopSuspendDeployment(namespace, deployment.Name)
		} else {
			suspend = k8s.suspendDeployment(ctx, namespace, deployment.Name)
		}

		s := kubesleep.NewSuspendable(
			kubesleep.Deplyoment,
			deployment.Name,
			*deployment.Spec.Replicas,
			suspend,
		)
		slog.Debug("parsed Suspendable", "Suspendable", s, "namespace", namespace)
		suspendables[s.Identifier()] = s
	}

	return suspendables, nil
}

func (k8s K8Simpl) noopSuspendDeployment(namespace, name string) func() error {
	return func() error {
		slog.Debug("Deployment already at 0 replicas; skipping suspend", "namespace", namespace, "name", name)
		return nil
	}
}

func (k8s K8Simpl) suspendDeployment(ctx context.Context, namespace string, name string) func() error {
	return func() error {
		scalable, err := k8s.clientset.AppsV1().
			Deployments(namespace).
			GetScale(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		scalable.Spec.Replicas = int32(0)

		_, err = k8s.clientset.AppsV1().Deployments(namespace).UpdateScale(
			ctx,
			scalable.Name,
			scalable,
			metav1.UpdateOptions{},
		)
		if err != nil {
			return err
		}
		slog.Info("Suspended Deployment", "name", scalable.Name, "namespace", namespace)
		return nil
	}
}

func (k8s K8Simpl) scaleDeployment(ctx context.Context, namespace string, name string, replicas int32) error {
	scalable, err := k8s.clientset.AppsV1().
		Deployments(namespace).
		GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	scalable.Spec.Replicas = replicas
	_, err = k8s.clientset.AppsV1().Deployments(scalable.Namespace).UpdateScale(
		ctx,
		scalable.Name,
		scalable,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}

	slog.Info("Woke up Deployment", "namespace", namespace, "name", name)
	return nil
}
