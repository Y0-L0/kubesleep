package k8s

import (
	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateDeployment(k8s K8Simpl, namespace string, name string, Replicas int32) (func() error, error) {
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

	_, err := k8s.clientset.AppsV1().Deployments(namespace).Create(k8s.ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	delete := func() error {
		return k8s.clientset.AppsV1().Deployments(namespace).Delete(k8s.ctx, name, metav1.DeleteOptions{})
	}
	return delete, nil
}

func (s *Integrationtest) TestCreateDeleteDeployment() {
	deleteNamespace, err := testNamespace("create-delete-deployment", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateDeployment(*s.k8s, "create-delete-deployment", "test-deployment", int32(2))
	s.Require().NoError(err)
	defer delete()
}

func (s *Integrationtest) TestGetDeployment() {
	deleteNamespace, err := testNamespace("get-deployments", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateDeployment(*s.k8s, "get-deployments", "test-deployment", int32(2))
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
	deleteNamespace, err := testNamespace("suspend-deployments-via-suspendable", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateDeployment(*s.k8s, "suspend-deployments-via-suspendable", "test-deployment", int32(2))
	s.Require().NoError(err)
	defer delete()

	before := s.getSuspendable("suspend-deployments-via-suspendable", "0:test-deployment")
	s.Require().Equal(int32(2), before.Replicas)

	s.Require().NoError(before.Suspend())

	actual := s.getSuspendable("suspend-deployments-via-suspendable", "0:test-deployment")
	s.Require().Equal(int32(0), actual.Replicas)
}

func (s *Integrationtest) TestScaleDeployment() {
	deleteNamespace, err := testNamespace("scale-deployments", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateDeployment(*s.k8s, "scale-deployments", "test-deployment", int32(0))
	s.Require().NoError(err)
	defer delete()

	err = s.k8s.ScaleSuspendable("scale-deployments", kubesleep.Deplyoment, "test-deployment", int32(2))

	actual := s.getSuspendable("scale-deployments", "0:test-deployment")
	s.Require().Equal(int32(2), actual.Replicas)
}
