package kubesleep

import (
	"log/slog"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s k8simpl) getDeployments(namespace string) (map[string]suspendable, error) {
	deployments, err := k8s.clientset.AppsV1().
		Deployments(namespace).
		List(k8s.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	suspendables := map[string]suspendable{}

	for _, deployment := range deployments.Items {
		suspend := func() error {
			return k8s.suspendDeployment(&deployment)
		}

		s := suspendable{
			manifestType: "Deployment",
			name:         deployment.ObjectMeta.Name,
			replicas:     *deployment.Spec.Replicas,
			suspend:      suspend,
		}
		slog.Debug("parsed suspendable", "suspendable", s)
		suspendables[s.Identifier()] = s
	}

	return suspendables, nil
}

func (k8s k8simpl) suspendDeployment(deployment *appsv1.Deployment) error {
	replicas := int32(0)
	deployment.Spec.Replicas = &replicas
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

func (k8s k8simpl) scaleDeployment(namespace string, name string, replicas int32) error {
	deployment, err := k8s.clientset.AppsV1().Deployments(namespace).Get(
		k8s.ctx,
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		return err
	}
	deployment.Spec.Replicas = &replicas

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
