package k8s

import (
	"log/slog"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	appsv1 "k8s.io/api/apps/v1"
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
		suspend := func() error {
			return k8s.suspendDeployment(&deployment)
		}

		s := kubesleep.NewSuspendable(
			kubesleep.Deplyoment,
			deployment.ObjectMeta.Name,
			*deployment.Spec.Replicas,
			suspend,
		)
		slog.Debug("parsed Suspendable", "Suspendable", s)
		suspendables[s.Identifier()] = s
	}

	return suspendables, nil
}

func (k8s K8Simpl) suspendDeployment(deployment *appsv1.Deployment) error {
	Replicas := int32(0)
	deployment.Spec.Replicas = &Replicas
	_, err := k8s.clientset.AppsV1().Deployments(deployment.Namespace).Update(
		k8s.ctx,
		deployment,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}
	slog.Info("Suspended Deployment", "name", deployment.Name)
	return nil
}

func (k8s K8Simpl) ScaleDeployment(namespace string, name string, Replicas int32) error {
	deployment, err := k8s.clientset.AppsV1().Deployments(namespace).Get(
		k8s.ctx,
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		return err
	}
	deployment.Spec.Replicas = &Replicas

	_, err = k8s.clientset.AppsV1().Deployments(deployment.Namespace).Update(
		k8s.ctx,
		deployment,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}

	slog.Info("Woke up Deployment", "name", name)
	return nil
}
