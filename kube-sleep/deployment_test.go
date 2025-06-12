package kubesleep

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createDeployment(k8s k8simpl, namespace string, name string, replicas int32) (func() error, error) {
	labels := map[string]string{"app": name}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
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
	deleteNamespace, err := testNamespace("create-delete-deployment", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := createDeployment(*s.k8s, "create-delete-deployment", "test-deployment", int32(2))
	s.Require().NoError(err)
	defer delete()
}

func (s *Integrationtest) TestGetDeployment() {
	deleteNamespace, err := testNamespace("get-deployments", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := createDeployment(*s.k8s, "get-deployments", "test-deployment", int32(2))
	s.Require().NoError(err)
	defer delete()

	sus, err := s.k8s.getDeployments("get-deployments")
	s.Require().NoError(err)
	s.Require().NotEmpty(sus)

	// simplify for easier comparison
	actual := sus["Deploymenttest-deployment"]
	actual.suspend = nil
	s.Require().Equal(
		suspendable{
			manifestType: "Deployment",
			name:         "test-deployment",
			replicas:     int32(2),
		},
		actual,
	)
}

func (s *Integrationtest) TestSuspendDeployment() {
	deleteNamespace, err := testNamespace("suspend-deployments", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := createDeployment(*s.k8s, "suspend-deployments", "test-deployment", int32(2))
	s.Require().NoError(err)
	defer delete()

	sus, err := s.k8s.getDeployments("suspend-deployments")
	s.Require().NoError(err)
	s.Require().NotEmpty(sus)
	s.Require().Equal(int32(2), sus["Deploymenttest-deployment"].replicas)

	err = sus["Deploymenttest-deployment"].suspend()
	s.Require().NoError(err)

	sus, err = s.k8s.getDeployments("suspend-deployments")
	s.Require().NoError(err)
	s.Require().NotEmpty(sus)
	s.Require().Equal(int32(0), sus["Deploymenttest-deployment"].replicas)
}

func (s *Integrationtest) TestScaleDeployment() {
	deleteNamespace, err := testNamespace("scale-deployments", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := createDeployment(*s.k8s, "scale-deployments", "test-deployment", int32(0))
	s.Require().NoError(err)
	defer delete()

	err = s.k8s.scaleDeployment("scale-deployments", "test-deployment", int32(2))

	sus, err := s.k8s.getDeployments("scale-deployments")
	s.Require().NoError(err)
	s.Require().NotEmpty(sus)
	s.Require().Equal(int32(2), sus["Deploymenttest-deployment"].replicas)
}
