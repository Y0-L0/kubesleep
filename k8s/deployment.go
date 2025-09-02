package k8s

import (
	"log/slog"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s K8Simpl) GetDeployments(namespace string) (map[string]kubesleep.Suspendable, error) {
	deployments, err := k8s.clientset.AppsV1().
		Deployments(namespace).
		List(k8s.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	suspendables := map[string]kubesleep.Suspendable{}

	for _, deployment := range deployments.Items {
		suspend := k8s.suspendDeployment(namespace, deployment.Name)

		s := kubesleep.NewSuspendable(
			kubesleep.Deplyoment,
			deployment.Name,
			*deployment.Spec.Replicas,
			suspend,
		)
		slog.Debug("parsed Suspendable", "Suspendable", s)
		suspendables[s.Identifier()] = s
	}

	return suspendables, nil
}

func (k8s K8Simpl) suspendDeployment(namespace string, name string) func() error {
	return func() error {
		scalable, err := k8s.clientset.AppsV1().
			Deployments(namespace).
			GetScale(k8s.ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		scalable.Spec.Replicas = int32(0)

		_, err = k8s.clientset.AppsV1().Deployments(namespace).UpdateScale(
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

func (k8s K8Simpl) ScaleDeployment(namespace string, name string, Replicas int32) error {
	scalable, err := k8s.clientset.AppsV1().
		Deployments(namespace).
		GetScale(k8s.ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	scalable.Spec.Replicas = Replicas
	_, err = k8s.clientset.AppsV1().Deployments(scalable.Namespace).UpdateScale(
		k8s.ctx,
		scalable.Name,
		scalable,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}

	slog.Info("Woke up Deployment", "name", name)
	return nil
}
