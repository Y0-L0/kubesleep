package k8s

import (
	"context"
	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateDeployment(ctx context.Context, k8s K8Simpl, namespace string, name string, Replicas int32) (func() error, error) {
	labels := map[string]string{"app": name}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: "k8s.gcr.io/pause:3.9",
						},
					},
				},
			},
		},
	}

	_, err := k8s.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	delete := func() error {
		return k8s.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	}
	return delete, nil
}

func (s *Integrationtest) TestCreateDeleteDeployment() {
	deleteNamespace, err := testNamespace(s.ctx, "create-delete-deployment", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateDeployment(s.ctx, *s.k8s, "create-delete-deployment", "test-deployment", int32(2))
	s.Require().NoError(err)
	defer delete()
}

func (s *Integrationtest) TestGetDeployment() {
	deleteNamespace, err := testNamespace(s.ctx, "get-deployments", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateDeployment(s.ctx, *s.k8s, "get-deployments", "test-deployment", int32(2))
	s.Require().NoError(err)
	defer delete()

	actual := s.getSuspendable("get-deployments", "0:test-deployment")
	actual.Suspend = nil
	s.Require().Equal(
		kubesleep.NewSuspendable(
			kubesleep.Deplyoment,
			"test-deployment",
			int32(2),
			nil,
		),
		actual,
	)
}

func (s *Integrationtest) TestSuspendDeployment() {
	deleteNamespace, err := testNamespace(s.ctx, "suspend-deployments-via-suspendable", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateDeployment(s.ctx, *s.k8s, "suspend-deployments-via-suspendable", "test-deployment", int32(2))
	s.Require().NoError(err)
	defer delete()

	before := s.getSuspendable("suspend-deployments-via-suspendable", "0:test-deployment")
	s.Require().Equal(int32(2), before.Replicas)

	s.Require().NoError(before.Suspend())

	actual := s.getSuspendable("suspend-deployments-via-suspendable", "0:test-deployment")
	s.Require().Equal(int32(0), actual.Replicas)
}

func (s *Integrationtest) TestAlreadySuspendedDeployment() {
	deleteNamespace, err := testNamespace(s.ctx, "skip-already-suspended-deployment", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateDeployment(s.ctx, *s.k8s, "skip-already-suspended-deployment", "test-deployment", int32(0))
	s.Require().NoError(err)
	defer delete()

	beforeScale, err := s.k8s.clientset.AppsV1().Deployments("skip-already-suspended-deployment").GetScale(s.ctx, "test-deployment", metav1.GetOptions{})
	s.Require().NoError(err)

	before := s.getSuspendable("skip-already-suspended-deployment", "0:test-deployment")
	s.Require().Equal(int32(0), before.Replicas)

	s.Require().NoError(before.Suspend())

	afterScale, err := s.k8s.clientset.AppsV1().Deployments("skip-already-suspended-deployment").GetScale(s.ctx, "test-deployment", metav1.GetOptions{})
	s.Require().NoError(err)
	s.Require().Equal(beforeScale.ResourceVersion, afterScale.ResourceVersion)
}

func (s *Integrationtest) TestScaleDeployment() {
	deleteNamespace, err := testNamespace(s.ctx, "scale-deployments", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateDeployment(s.ctx, *s.k8s, "scale-deployments", "test-deployment", int32(0))
	s.Require().NoError(err)
	defer delete()

	err = s.k8s.ScaleSuspendable(s.ctx, "scale-deployments", kubesleep.Deplyoment, "test-deployment", int32(2))

	actual := s.getSuspendable("scale-deployments", "0:test-deployment")
	s.Require().Equal(int32(2), actual.Replicas)
}
